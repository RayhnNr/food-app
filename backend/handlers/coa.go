package handlers

import (
	"fmt"
	"food-api/database"
	"food-api/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const detailSuffix = ",00"

type CoaInput struct {
	CoaNo     string `json:"coa_no"`
	CoaName   string `json:"coa_name"`
	TypeID    uint   `json:"coa_type_id"`
	CoaParent *uint  `json:"coa_parent"`
	CoaHead   bool   `json:"coa_head"`
}

type CoaNode struct {
	CoaID    uint       `json:"coa_id"`
	CoaNo    string     `json:"coa_no"`
	CoaName  string     `json:"coa_name"`
	TypeID   uint       `json:"coa_type_id"`
	TypeName string     `json:"coa_type_name"`
	Parent   *uint      `json:"coa_parent"`
	Head     bool       `json:"coa_head"`
	Children []*CoaNode `json:"children"`
}

func validasiCoa(in *CoaInput, id uint) map[string]string {
	err := map[string]string{}

	in.CoaNo = strings.TrimSpace(in.CoaNo)
	in.CoaName = strings.TrimSpace(in.CoaName)

	if in.CoaNo == "" {
		err["coa_no"] = "Kode akun wajib diisi"
	} else {
		var n int64
		database.DB.Model(&models.Coa{}).Where("coa_no = ? AND coa_id <> ?", in.CoaNo, id).Count(&n)
		if n > 0 {
			err["coa_no"] = "Kode akun sudah dipakai"
		}
	}

	if in.CoaName == "" {
		err["coa_name"] = "Nama akun wajib diisi"
	} else if len([]rune(in.CoaName)) < 3 {
		err["coa_name"] = "Nama akun minimal 3 karakter"
	}

	if in.TypeID == 0 {
		err["coa_type_id"] = "Tipe akun wajib diisi"
	} else {
		var n int64
		database.DB.Model(&models.CoaType{}).Where("coa_type_id = ?", in.TypeID).Count(&n)
		if n == 0 {
			err["coa_type_id"] = "Tipe akun tidak ditemukan"
		}
	}

	if in.CoaParent != nil {
		if id != 0 && *in.CoaParent == id {
			err["coa_parent"] = "Akun tidak boleh menjadi parent dirinya sendiri"
		} else {
			var p models.Coa
			if e := database.DB.First(&p, "coa_id = ?", *in.CoaParent).Error; e != nil {
				err["coa_parent"] = "Parent tidak ditemukan"
			} else if !p.CoaHead {
				err["coa_parent"] = "Parent harus akun header"
			} else if id != 0 && isDescendant(id, *in.CoaParent) {
				err["coa_parent"] = "Parent tidak boleh turunan akun ini sendiri"
			}
		}
	}

	return err
}

// isDescendant mengecek apakah kandidat berada di bawah akun id.
func isDescendant(id uint, kandidat uint) bool {
	cur := kandidat
	for i := 0; i < 50; i++ {
		if cur == id {
			return true
		}

		var p models.Coa
		if err := database.DB.Select("coa_id", "coa_parent").First(&p, "coa_id = ?", cur).Error; err != nil {
			return false
		}
		if p.CoaParent == nil {
			return false
		}

		cur = *p.CoaParent
	}

	// kedalaman melebihi batas, anggap loop dan tolak
	return true
}

func buatTree(coas []models.Coa) []*CoaNode {
	nodes := make(map[uint]*CoaNode, len(coas))
	for _, c := range coas {
		n := &CoaNode{
			CoaID:    c.CoaID,
			CoaNo:    c.CoaNo,
			CoaName:  c.CoaName,
			TypeID:   c.TypeID,
			Parent:   c.CoaParent,
			Head:     c.CoaHead,
			Children: []*CoaNode{},
		}
		if c.Type != nil {
			n.TypeName = c.Type.CoaTypeName
		}
		nodes[c.CoaID] = n
	}

	roots := []*CoaNode{}
	for _, c := range coas {
		n := nodes[c.CoaID]
		if c.CoaParent == nil {
			roots = append(roots, n)
		} else if p, ok := nodes[*c.CoaParent]; ok {
			p.Children = append(p.Children, n)
		} else {
			roots = append(roots, n)
		}
	}

	return roots
}

func parseNo(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, detailSuffix) {
		s = strings.TrimSuffix(s, detailSuffix)
	}
	if strings.Contains(s, ",") {
		return 0, false
	}
	digits := strings.ReplaceAll(s, ".", "")
	if len(digits) != 9 {
		return 0, false
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func formatCoaNo(n int64) string {
	return fmt.Sprintf("%03d.%03d.%03d", n/1000000, (n/1000)%1000, n%1000)
}

func coaNoExists(no string) (bool, error) {
	base := strings.TrimSuffix(no, detailSuffix)
	var count int64
	err := database.DB.Model(&models.Coa{}).Where("coa_no IN ?", []string{base, base + detailSuffix}).Count(&count).Error
	return count > 0, err
}

func hitungKodeCoaBerikutnya(parentNo string, childCodes []string, isHeader bool) (string, error) {
	var next, step, limit int64

	if parentNo == "" {
		if !isHeader {
			return "", fmt.Errorf("Akun level 1 harus berupa header")
		}
		next, step, limit = 100000000, 1000000, 1000000000
		for _, childNo := range childCodes {
			n, ok := parseNo(childNo)
			if !ok || n%1000000 != 0 || n < 100000000 || n >= limit {
				continue
			}
			if candidate := (n/1000000 + 1) * 1000000; candidate > next {
				next = candidate
			}
		}
	} else {
		base, ok := parseNo(parentNo)
		if !ok {
			return "", fmt.Errorf("Kode parent tidak valid, isi kode manual")
		}
		if base%1000 != 0 {
			return "", fmt.Errorf("Akun detail tidak boleh memiliki anak")
		}

		isLevel1 := base%1000000 == 0
		if isHeader && !isLevel1 {
			return "", fmt.Errorf("Header anak tidak boleh memiliki header lagi")
		}

		if isHeader {
			next, step, limit = base+100000, 100000, base+1000000
			for _, childNo := range childCodes {
				n, ok := parseNo(childNo)
				if !ok || n <= base || n >= limit || n%1000 != 0 {
					continue
				}
				segment := (n - base) / 1000
				if segment >= 100 && segment <= 999 {
					if candidate := n + step; candidate > next {
						next = candidate
					}
				}
			}
		} else {
			next, step, limit = base+1, 1, base+1000
			for _, childNo := range childCodes {
				n, ok := parseNo(childNo)
				if ok && n > base && n < limit && n >= next {
					next = n + 1
				}
			}
		}
	}

	for next < limit && kodeCoaTerpakai(childCodes, next) {
		next += step
	}
	if next >= limit {
		return "", fmt.Errorf("Kode sudah penuh, isi kode manual")
	}

	no := formatCoaNo(next)
	if !isHeader {
		no += detailSuffix
	}
	return no, nil
}

func kodeCoaTerpakai(codes []string, candidate int64) bool {
	for _, code := range codes {
		if n, ok := parseNo(code); ok && n == candidate {
			return true
		}
	}
	return false
}

func GetCoaTypes(c *gin.Context) {
	types := []models.CoaType{}

	if err := database.DB.Order("coa_type_id").Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data tipe COA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": types})
}

func GetCoa(c *gin.Context) {
	coas := []models.Coa{}

	if err := database.DB.Preload("Type").Order("coa_no").Find(&coas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data COA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": coas})
}

func GetCoaTree(c *gin.Context) {
	coas := []models.Coa{}

	if err := database.DB.Preload("Type").Order("coa_no").Find(&coas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data COA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": buatTree(coas)})
}

// GET /coa/next-code?parent_id=5&head=true
// tanpa parent_id = header level 1, head=false = akun detail
func GetNextCoaCode(c *gin.Context) {
	isHeader := c.Query("head") != "false"
	var parentNo string
	var children []models.Coa
	if pid := c.Query("parent_id"); pid == "" {
		if err := database.DB.Select("coa_no").Where("coa_parent IS NULL").Find(&children).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data COA"})
			return
		}
	} else {
		var parent models.Coa
		if err := database.DB.First(&parent, "coa_id = ?", pid).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Parent tidak ditemukan"})
			return
		}
		parentNo = parent.CoaNo
		if err := database.DB.Select("coa_no").Where("coa_parent = ?", parent.CoaID).Find(&children).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data COA"})
			return
		}
	}

	childCodes := make([]string, 0, len(children))
	for _, child := range children {
		childCodes = append(childCodes, child.CoaNo)
	}
	var no string
	for {
		var err error
		no, err = hitungKodeCoaBerikutnya(parentNo, childCodes, isHeader)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
			return
		}
		used, err := coaNoExists(no)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal memeriksa kode COA"})
			return
		}
		if !used {
			break
		}
		childCodes = append(childCodes, no)
	}
	c.JSON(http.StatusOK, gin.H{"coa_no": no})
}

func GetCoaBy(c *gin.Context) {
	var coa models.Coa

	if err := database.DB.Preload("Type").First(&coa, "coa_id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "COA tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, coa)
}

func CreateCoa(c *gin.Context) {
	var in CoaInput

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "JSON tidak valid"})
		return
	}

	if errs := validasiCoa(&in, 0); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	coa := models.Coa{
		CoaNo:     in.CoaNo,
		CoaName:   in.CoaName,
		TypeID:    in.TypeID,
		CoaParent: in.CoaParent,
		CoaHead:   in.CoaHead,
	}

	if err := database.DB.Create(&coa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan data"})
		return
	}

	c.JSON(http.StatusCreated, coa)
}

func UpdateCoa(c *gin.Context) {
	var coa models.Coa

	if err := database.DB.First(&coa, "coa_id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "COA tidak ditemukan"})
		return
	}

	var in CoaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "JSON tidak valid"})
		return
	}

	errs := validasiCoa(&in, coa.CoaID)

	if coa.CoaHead && !in.CoaHead {
		var n int64
		database.DB.Model(&models.Coa{}).Where("coa_parent = ?", coa.CoaID).Count(&n)
		if n > 0 {
			errs["coa_head"] = "Akun masih punya anak, tidak bisa diubah menjadi akun detail"
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	coa.CoaNo = in.CoaNo
	coa.CoaName = in.CoaName
	coa.TypeID = in.TypeID
	coa.CoaParent = in.CoaParent
	coa.CoaHead = in.CoaHead

	if err := database.DB.Save(&coa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan data"})
		return
	}

	c.JSON(http.StatusOK, coa)
}

func DeleteCoa(c *gin.Context) {
	id := c.Param("id")

	var n int64
	database.DB.Model(&models.Coa{}).Where("coa_parent = ?", id).Count(&n)
	if n > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Akun masih punya anak, hapus anaknya dulu"})
		return
	}

	result := database.DB.Delete(&models.Coa{}, "coa_id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menghapus"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "COA tidak ditemukan"})
		return
	}

	c.Status(http.StatusNoContent)
}

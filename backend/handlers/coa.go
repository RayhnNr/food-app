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

type CoaInput struct {
	CoaNo     string `json:"coa_no"`
	CoaName   string `json:"coa_name"`
	TypeID    uint   `json:"coa_type_id"`
	CoaParent *uint  `json:"coa_parent"`
	CoaHead   bool   `json:"coa_head"`
}

// Validasi
func validasiCoa(in *CoaInput, id uint) map[string]string {
	err := map[string]string{}

	in.CoaNo = strings.TrimSpace(in.CoaNo)
	in.CoaName = strings.TrimSpace(in.CoaName)

	// Validasi coa_no
	if in.CoaNo == "" {
		err["coa_no"] = "Kode akun wajib diisi"
	} else {
		var n int64
		database.DB.Model(&models.Coa{}).Where("coa_no = ? AND coa_id <> ?", in.CoaNo, id).Count(&n)

		if n > 0 {
			err["coa_no"] = "Kode akun sudah dipakai"
		}
	}

	// Validasi coa_name
	if in.CoaName == "" {
		err["coa_name"] = "Nama akun wajib diisi"
	} else if len([]rune(in.CoaName)) < 3 {
		err["coa_name"] = "Nama akun minimal 3 karakter"
	}

	// Validasi coa_no
	if in.TypeID == 0 {
		err["coa_type_id"] = "Tipe akun wajib diisi"
	} else {
		var n int64
		database.DB.Model(&models.CoaType{}).Where("coa_type_id = ?", in.TypeID).Count(&n)

		if n == 0 {
			err["coa_type_id"] = "Tipe akun tidak ditemukan"
		}
	}

	// Validasi Coa Parent
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

func isDescendant(id uint, kandidat uint) bool {
	cur := kandidat
	for i := 0; i < 50; i++ {
		var p models.Coa
		if err := database.DB.Select("coa_id", "coa_parent").First(&p, "coa_id = ?", cur).Error; err != nil {
			return false
		}

		if *p.CoaParent == id {
			return true
		}

		cur = *p.CoaParent
	}

	return true
}

// Tree
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

func buatTree(coas []models.Coa) []*CoaNode {
	nodes := make(map[uint]*CoaNode, len(coas))
	for _, c := range coas {
		n := &CoaNode{
			CoaID: c.CoaID,
			CoaNo: c.CoaNo,
			CoaName: c.CoaName,
			TypeID: c.TypeID,
			Parent: c.CoaParent,
			Head: c.CoaHead,
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


// Generate Code otomatis
func parseNo(s string) (int64, bool) {
	n, err := strconv.ParseInt(strings.ReplaceAll(s, ".",""), 10, 64)
	if err != nil || n < 0 || n > 999999999 {
		return 0, false
	}
	return n, true
}

func formatCoaNo(n int64) string {
	return fmt.Sprintf("%03d.%03d.%03d", n/1000000, (n/1000)%1000, n%1000)
}

func coaNoExists(no string) bool {
	var n int64
	database.DB.Model(&models.Coa{}).Where("coa_no = ?", no).Count(&n)
	return n > 0
}

func GetCoaTypes(c *gin.Context) {
	types := []models.CoaType{}

	if err := database.DB.Order("coa_type_id").Find(&types).Error; err!= nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data tipe COA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": types})
}

// GET /coa
func GetCoa(c *gin.Context) {
	coas := []models.Coa{}
 
	if err := database.DB.Preload("Type").Order("coa_no").Find(&coas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data COA"})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{"data": coas})
}
 
// GET /coa/tree
func GetCoaTree(c *gin.Context) {
	coas := []models.Coa{}
 
	if err := database.DB.Preload("Type").Order("coa_no").Find(&coas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data COA"})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{"data": buatTree(coas)})
}
 
// GET /coa/next-code?parent_id=5 (tanpa parent_id = header paling atas)
func GetNextCoaCode(c *gin.Context) {
	pid := c.Query("parent_id")
	var next int64
 
	if pid != "" {
		var parent models.Coa
		if err := database.DB.First(&parent, "coa_id = ?", pid).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Parent tidak ditemukan"})
			return
		}
		base, ok := parseNo(parent.CoaNo)
		if !ok {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Kode parent tidak berformat 000.000.000, isi kode manual",
			})
			return
		}
 
		// naik dari anak dengan kode terbesar
		next = base
		var children []models.Coa
		database.DB.Select("coa_no").Where("coa_parent = ?", parent.CoaID).Find(&children)
		for _, ch := range children {
			if n, ok := parseNo(ch.CoaNo); ok && n > next {
				next = n
			}
		}
		next++
	} else {
		// header paling atas: naikkan segmen pertama
		var roots []models.Coa
		database.DB.Select("coa_no").Where("coa_parent IS NULL").Find(&roots)
 
		maxSeg := int64(99) // kode pertama jadi 100.000.000
		for _, r := range roots {
			if n, ok := parseNo(r.CoaNo); ok && n/1000000 > maxSeg {
				maxSeg = n / 1000000
			}
		}
		next = (maxSeg + 1) * 1000000
	}
 
	// lewati kode yang ternyata sudah dipakai
	step := int64(1)
	if pid == "" {
		step = 1000000
	}
	for i := 0; i < 1000 && coaNoExists(formatCoaNo(next)); i++ {
		next += step
	}
 
	if next > 999999999 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Kode sudah penuh, isi kode manual"})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{"coa_no": formatCoaNo(next)})
}
 
// GET /coa/:id
func GetCoaBy(c *gin.Context) {
	var coa models.Coa
 
	if err := database.DB.Preload("Type").First(&coa, "coa_id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "COA tidak ditemukan"})
		return
	}
 
	c.JSON(http.StatusOK, coa)
}
 
// POST /coa
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
 
// PUT /coa/:id
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
 
	// header yang masih punya anak tidak boleh jadi akun detail
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
 
// DELETE /coa/:id
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
 
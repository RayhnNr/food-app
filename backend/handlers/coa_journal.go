package handlers

import (
	// "context"
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
 
type CoaJournalItem struct {
	TypeID uint  `json:"id_type_coa"`
	CoaID  *uint `json:"coa_id"`
}
 
type CoaJournalInput struct {
	Items []CoaJournalItem `json:"items"`
}
 
type CoaJournalRow struct {
	TypeID  uint   `json:"id_type_coa"`
	Name    string `json:"name_coa_journal"`
	CoaID   *uint  `json:"coa_id"`
	CoaNo   string `json:"coa_no"`
	CoaName string `json:"coa_name"`
}

func currentUserID(c *gin.Context) (uint, bool) {
	return middlewares.GetUserIDFromCOntext(c)
}


func loadCoaJournal(userID uint) ([]CoaJournalRow, error) {
	rows := []CoaJournalRow{}

	err := database.DB.Table("coa_journal_type AS t").
		Select(`t.id_type_coa AS type_id,
			t.name_coa_journal AS name,
			j.coa_id AS coa_id,
			COALESCE(c.coa_no, '') AS coa_no,
			COALESCE(c.coa_name, '') AS coa_name`).
		Joins("LEFT JOIN coa_journal AS j ON j.id_type_coa = t.id_type_coa AND j.user_id = ?", userID).
		Joins("LEFT JOIN coas AS c ON c.coa_id = j.coa_id").
		Order("t.id_type_coa").
		Scan(&rows).Error
	return rows, err
}

func validasiCoaJournal(items []CoaJournalItem) map[string]string {
	errs := map[string]string{}
 
	if len(items) == 0 {
		errs["items"] = "Data tidak boleh kosong"
		return errs
	}
 
	typeIDs := make([]uint, 0, len(items))
	coaIDs := []uint{}
	seen := map[uint]bool{}
 
	for _, it := range items {
		key := strconv.FormatUint(uint64(it.TypeID), 10)
		if seen[it.TypeID] {
			errs[key] = "Tipe dikirim lebih dari satu kali"
			continue
		}
		seen[it.TypeID] = true
 
		typeIDs = append(typeIDs, it.TypeID)
		if it.CoaID != nil {
			coaIDs = append(coaIDs, *it.CoaID)
		}
	}
 
	validTypes := map[uint]bool{}
	var types []models.CoaJournalType
	database.DB.Select("id_type_coa").Where("id_type_coa IN ?", typeIDs).Find(&types)
	for _, t := range types {
		validTypes[t.IDTypeCoa] = true
	}
 
	validCoas := map[uint]bool{}
	if len(coaIDs) > 0 {
		var coas []models.Coa
		database.DB.Select("coa_id").Where("coa_id IN ? AND coa_no LIKE ?", coaIDs, "%,00").Find(&coas)
		for _, co := range coas {
			validCoas[co.CoaID] = true
		}
	}
 
	for _, it := range items {
		key := strconv.FormatUint(uint64(it.TypeID), 10)
		if _, bad := errs[key]; bad {
			continue
		}
 
		if !validTypes[it.TypeID] {
			errs[key] = "Tipe jurnal tidak ditemukan"
		} else if it.CoaID != nil && !validCoas[*it.CoaID] {
			errs[key] = "Akun tidak ditemukan atau kode akun tidak berakhiran ,00"
		}
	}
 
	return errs
}

func GetCoaJurnal(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	rows, err := loadCoaJournal(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal ambil data CoA jurnal",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : rows,
	})
}

func SaveCoaJournal(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	var in CoaJournalInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	for i := range in.Items {
		if in.Items[i].CoaID != nil && *in.Items[i].CoaID == 0 {
			in.Items[i].CoaID = nil
		}
	}

	if errs := validasiCoaJournal(in.Items); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, it := range in.Items {
			if it.CoaID == nil {
				err := tx.Where("user_id = ? AND id_type_coa = ?", userID, it.TypeID).
					Delete(&models.CoaJournal{}).Error
				if err != nil {
					return err
				}
				continue
			}
 
			row := models.CoaJournal{
				UserID: userID, 
				TypeID: it.TypeID, 
				IDCoa: *it.CoaID,
			}
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{
					{Name: "user_id"}, 
					{Name: "id_type_coa"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"coa_id", "modified"}),
			}).Create(&row).Error
			if err != nil {
				return err
			}
		}
 
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan CoA Journal",
		})
		return
	}

	rows, err := loadCoaJournal(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{
		"data": rows,
	})

}
package org

import (
	"gorm.io/gorm"
)

// Country 国家信息
type Country struct {
	gorm.Model
	Name string `json:"name" gorm:"not null;unique;size:100"`
	Code string `json:"code" gorm:"not null;unique;size:10"` // ISO 国家代码 (如：CN, US)
}

// TableName 指定表名
func (Country) TableName() string {
	return "countries"
}

// Agency 机构信息
type Agency struct {
	gorm.Model
	Name      string `json:"name" gorm:"not null;size:200"`
	CountryID uint   `json:"country_id" gorm:"not null;index"`
	Country   Country `json:"country" gorm:"foreignKey:CountryID"`
}

// TableName 指定表名
func (Agency) TableName() string {
	return "agencies"
}

// SeedData 初始化样例数据
func SeedData(db *gorm.DB) error {
	countries := []Country{
		{Name: "中国", Code: "CN"},
		{Name: "美国", Code: "US"},
		{Name: "日本", Code: "JP"},
		{Name: "德国", Code: "DE"},
		{Name: "法国", Code: "FR"},
		{Name: "英国", Code: "GB"},
	}

	for _, country := range countries {
		var existingCountry Country
		if err := db.Where("code = ?", country.Code).First(&existingCountry).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&country).Error; err != nil {
				return err
			}
		}
	}

	var cnCountry Country
	if err := db.Where("code = ?", "CN").First(&cnCountry).Error; err != nil {
		return err
	}

	agencies := []Agency{
		{Name: "国家发展和改革委员会", CountryID: cnCountry.ID},
		{Name: "中国生态文明研究与促进会", CountryID: cnCountry.ID},
		{Name: "中华环境保护基金会", CountryID: cnCountry.ID},
		{Name: "中国环境保护产业协会", CountryID: cnCountry.ID},
		{Name: "自然资源部", CountryID: cnCountry.ID},
		{Name: "生态环境部", CountryID: cnCountry.ID},
		{Name: "国家林业和草原局", CountryID: cnCountry.ID},
		{Name: "中国气候变化事务特使办公室", CountryID: cnCountry.ID},
	}

	for _, agency := range agencies {
		var existingAgency Agency
		if err := db.Where("name = ? AND country_id = ?", agency.Name, agency.CountryID).First(&existingAgency).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&agency).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

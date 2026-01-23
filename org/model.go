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
}

// TableName 指定表名
func (Agency) TableName() string {
	return "agencies"
}

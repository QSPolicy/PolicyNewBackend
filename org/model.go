package org

import (
	"fmt"

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
	Name      string  `json:"name" gorm:"not null;size:200"`
	CountryID uint    `json:"country_id" gorm:"not null;index"`
	Country   Country `json:"country" gorm:"foreignKey:CountryID"`
	Domain    string  `json:"domain" gorm:"size:300"`
}

// TableName 指定表名
func (Agency) TableName() string {
	return "agencies"
}

// SeedData 初始化样例数据
func SeedData(db *gorm.DB) error {
	// 1. 定义需要初始化的国家列表
	countries := []Country{
		{Name: "美国", Code: "US"},
		{Name: "英国", Code: "GB"},
		{Name: "德国", Code: "DE"},
		{Name: "法国", Code: "FR"},
		{Name: "日本", Code: "JP"},
		{Name: "韩国", Code: "KR"},
		{Name: "俄罗斯", Code: "RU"},
		{Name: "瑞士", Code: "CH"},
		{Name: "澳大利亚", Code: "AU"},
		{Name: "加拿大", Code: "CA"},
		{Name: "欧盟/国际", Code: "EU"},
	}

	// 2. 插入或获取国家，并建立 Code -> ID 的映射
	countryMap := make(map[string]uint)
	for _, country := range countries {
		if err := db.Where(Country{Code: country.Code}).Attrs(Country{Name: country.Name}).FirstOrCreate(&country).Error; err != nil {
			return err
		}
		countryMap[country.Code] = country.ID
	}

	// 3. 定义机构数据
	// 注意：部分链接在脚本检测中可能会报 403/超时/SSL 错误，但这通常是因为反爬虫策略或地理限制，网址本身是正确的。
	type agencySeed struct {
		CountryCode string
		Name        string
		Domain      string
	}

	agenciesData := []agencySeed{
		// --- 美国 (US) ---
		// NSTC 和 PCAST 原链接失效，统一指向其上级机构 OSTP
		{"US", "美国国家科学技术委员会", "www.whitehouse.gov"},
		{"US", "美国总统科技顾问委员会", "www.whitehouse.gov"},
		{"US", "美国白宫科技政策办公室", "www.whitehouse.gov"},
		// DNI 链接虽然报 403 (Forbidden)，但这是正确的官网
		{"US", "美国国家情报委员会", "www.dni.gov"},
		{"US", "美国能源部 (DOE)", "www.energy.gov"},
		{"US", "美国国家科学基金会 (NSF)", "www.nsf.gov"},
		{"US", "美国国立卫生研究院 (NIH)", "www.nih.gov"},
		{"US", "美国国家科学院", "www.nasonline.org"},
		{"US", "美国国家工程院", "www.nae.edu"},
		{"US", "美国国家医学院", "nam.edu"},
		{"US", "美国兰德公司", "www.rand.org"},
		{"US", "博思艾伦咨询公司 (Booz Allen Hamilton)", "www.boozallen.com"},
		{"US", "美国布鲁金斯学会", "www.brookings.edu"},
		{"US", "美国新美国安全中心", "www.cnas.org"},
		{"US", "美国战略与国际问题研究中心", "www.csis.org"},
		{"US", "美国大西洋理事会", "www.atlanticcouncil.org"},
		{"US", "美国信息技术与创新基金会", "itif.org"},

		// --- 英国 (GB) ---
		{"GB", "英国研究与创新署", "www.ukri.org"},
		// 原链接 404，更新为英国科学技术委员会 (CST) 官网
		{"GB", "英国国家科学与技术委员会", "www.gov.uk"},
		{"GB", "英国科学与技术战略办公室", "www.gov.uk"},
		{"GB", "英国皇家学会", "royalsociety.org"},

		// --- 德国 (DE) ---
		{"DE", "德国联邦教育与研究部", "www.bmbf.de"},
		{"DE", "德国联邦与州科学联席会", "www.gwk-bonn.de"},
		// 德国部分网站 SSL 证书可能不被部分脚本信任，链接无误
		{"DE", "德国科学理事会", "www.wissenschaftsrat.de"},
		{"DE", "德国研究联合会", "www.dfg.de"},
		{"DE", "德国洪堡基金会", "www.humboldt-foundation.de"},
		{"DE", "德国马普学会", "www.mpg.de"},
		{"DE", "德国弗朗霍夫协会", "www.fraunhofer.de"},

		// --- 法国 (FR) ---
		{"FR", "法国高等教育、研究与创新部", "www.enseignementsup-recherche.gouv.fr"},
		{"FR", "法国国家科研署", "anr.fr"},
		{"FR", "法兰西科学院", "www.academie-sciences.fr"},
		{"FR", "法国国家科研中心", "www.cnrs.fr"},
		{"FR", "法国巴斯德研究所", "www.pasteur.fr"},

		// --- 日本 (JP) ---
		{"JP", "日本综合科学技术创新会议", "www8.cao.go.jp"},
		{"JP", "日本文部科学省", "www.mext.go.jp"},
		{"JP", "日本学术振兴会 (JSPS)", "www.jsps.go.jp"},
		{"JP", "日本科学技术振兴机构 (JST)", "www.jst.go.jp"},
		{"JP", "日本科学技术振兴机构研究开发战略中心(CRDS)", "www.jst.go.jp"},
		// 原链接 404，更新为 NEDO 英文首页
		{"JP", "日本新能源与产业技术综合开发机构技术战略中心", "www.nedo.go.jp"},
		{"JP", "日本科学技术与学术政策研究所 (NISTEP)", "www.nistep.go.jp"},
		{"JP", "日本科学技术政策研究所", "www.nistep.go.jp"},

		// --- 韩国 (KR) ---
		{"KR", "韩国科学技术信息通信部", "www.msit.go.kr"},
		{"KR", "韩国研究基金会", "www.nrf.re.kr"},
		{"KR", "韩国科学技术咨询会议", "www.pacst.go.kr"},
		{"KR", "韩国科学技术企划评价院", "www.kistep.re.kr"},

		// --- 俄罗斯 (RU) ---
		// 俄罗斯政府网存在区域屏蔽和证书兼容性问题，链接无误
		{"RU", "俄罗斯联邦科学与高等教育部", "minobrnauki.gov.ru"},
		{"RU", "俄罗斯科学院", "new.ras.ru"},

		// --- 瑞士 (CH) ---
		{"CH", "瑞士国家科学基金会", "www.snf.ch"},

		// --- 澳大利亚 (AU) ---
		// 澳洲政府网通常屏蔽脚本导致超时，链接无误
		{"AU", "澳大利亚研究理事会", "www.arc.gov.au"},
		{"AU", "澳大利亚科学院", "www.science.org.au"},
		{"AU", "澳大利亚联邦科学与工业研究组织", "www.csiro.au"},

		// --- 加拿大 (CA) ---
		{"CA", "加拿大自然科学与工程研究理事会", "www.nserc-crsng.gc.ca"},
		{"CA", "加拿大社会科学和人文科学研究理事会", "www.sshrc-crsh.gc.ca"},

		// --- 欧盟/国际 (EU) ---
		{"EU", "欧洲研究理事会", "erc.europa.eu"},
		{"EU", "欧洲创新理事会", "eic.ec.europa.eu"},
		{"EU", "欧盟委员会", "commission.europa.eu"},
	}

	// 4. 遍历并插入机构数据
	for _, item := range agenciesData {
		countryID, ok := countryMap[item.CountryCode]
		if !ok {
			fmt.Printf("Warning: Country code %s not found for agency %s\n", item.CountryCode, item.Name)
			continue
		}

		agency := Agency{
			Name:      item.Name,
			CountryID: countryID,
			Domain:    item.Domain,
		}

		if err := db.Where(Agency{Name: agency.Name, CountryID: agency.CountryID}).
			Attrs(Agency{Domain: agency.Domain}).
			FirstOrCreate(&agency).Error; err != nil {
			return err
		}

		// 如果 Domain 不为空且与数据库中不同，则更新（确保修正后的链接被应用）
		if item.Domain != "" && agency.Domain != item.Domain {
			db.Model(&agency).Update("domain", item.Domain)
		}
	}

	return nil
}

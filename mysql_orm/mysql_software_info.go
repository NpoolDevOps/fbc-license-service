package mysql_orm


import (
    "time"
    "github.com/jinzhu/gorm"
)


type SoftwareInfo struct{
    Id string            `gorm:"column:id;primary_key"`
    SoftwareSn string    `gorm:"column:software_sn"`
    SystemSn   string    `gorm:"column:system_sn"`
    Status     int       `gorm:"column:status"`
    DevopsStatus int     `gorm:"column:devops_status"`
    CreateTime time.Time    `gorm:"column:create_time"`
    ModifyTime time.Time    `gorm:"column:modify_time"`
}


func InsertSoftwareInfo(db *gorm.DB, softwareInfo SoftwareInfo){
    db.Create(&softwareInfo)
}


func QuerySoftwareInfoBySystemSn(db *gorm.DB, systemSn string) *SoftwareInfo{

    var softInfo SoftwareInfo
    var count int
    db.Where("system_sn = ?", systemSn).Find(&softInfo).Count(&count)

    if count == 0{
        return nil
    }
    
    return &softInfo
}


func GetSoftwareCount(db *gorm.DB, softwareSn string) int {
    
    var softInfos []SoftwareInfo
    var count int

    db.Where("software_sn = ?", softwareSn).Find(&softInfos).Count(&count)
    return count
}


func QuerySoftwareInfos(db *gorm.DB) [] SoftwareInfo{

    var softInfos []SoftwareInfo
    db.Find(&softInfos)
    return softInfos
}


func GetSoftwareDevopsStatus(db *gorm.DB, uuid string) SoftwareInfo {

    var softInfo SoftwareInfo
    var count int
    db.Find(&softInfo).Count(&count)

    if count == 0{
        return nil
    }
    return softInfo
}


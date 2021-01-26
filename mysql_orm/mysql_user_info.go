package mysql_orm


import (
    "github.com/jinzhu/gorm"
    "time"
)


type UserInfo struct{

    Id string     `gorm:"column:id;primary_key"`
    UserName string     `gorm:"column:user_name"`
    ValidateDate   time.Time     `gorm:"column:validate_date"`
    Volume    int   `gorm:"column:volume"`
    Sn        string `gorm:"column:sn"`
    CreateTime   time.Time    `gorm:"column:create_time"`
    ModifyTime   time.Time    `gorm:"column:modify_time"`
}


func QueryUserInfoBySn(db *gorm.DB, softwareSn string) *UserInfo{

    var userInfo UserInfo
    var count int
    db.Where("sn = ?", softwareSn).Find(&userInfo).Count(&count)
    if count == 0{
        return nil
    }

    return &userInfo
}


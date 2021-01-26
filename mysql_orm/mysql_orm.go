package mysql_orm

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
)


func NewDbOrm(dbType string, connectUrl string)(*gorm.DB, error){
    db, err := gorm.Open(dbType, connectUrl)
    if err != nil{
        return nil, err
    }
    db.SingularTable(true)
    return db, nil
}


func DbOrmClose(db *gorm.DB){
    db.Close()
}


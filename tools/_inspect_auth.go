package main
import (
 "database/sql"
 "fmt"
 _ "modernc.org/sqlite"
)
func main(){ db,err:=sql.Open("sqlite","./db/chatnio.db"); if err!=nil{panic(err)}; defer db.Close(); rows,err:=db.Query("SELECT id, username, email, is_admin FROM auth ORDER BY id"); if err!=nil{panic(err)}; defer rows.Close(); for rows.Next(){ var id int; var u,e string; var a any; if err:=rows.Scan(&id,&u,&e,&a); err!=nil{panic(err)}; fmt.Printf("id=%d username=%s email=%s is_admin=%v\n",id,u,e,a)} }

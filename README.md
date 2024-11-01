package main

import (
	"fmt"
	"log"
	"github.com/GitbookIO/go-gitbook-api/api"
	"github.com/GitbookIO/go-gitbook-api/client"
)

// ฟังก์ชันหลัก
func main() {
	// ตั้งค่า APIOptions สำหรับ Client ของคุณ
	opts := client.ClientOptions{
		BaseURL: "https://api.gitbook.com", // URL ของ Gitbook API
	}

	// สร้างอินสแตนซ์ของ API
	apiClient := NewAPI(APIOptions(opts))

	// ตัวอย่างการเรียก Book.Get (คุณต้องกำหนด Book ID หรือ slug ที่ต้องการ)
	bookID := "Jiraphat3112"
	book, err := apiClient.Book.Get(bookID)
	if err != nil {
		log.Fatal("ไม่สามารถเรียกข้อมูลหนังสือได้:", err)
	}

	// แสดงข้อมูลของหนังสือ
	fmt.Println("ข้อมูลหนังสือ:", book)
}
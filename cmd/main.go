package main

import (
	"fmt"
	"fuzzyTest/db"
	"fuzzyTest/medName"
	"fuzzyTest/services"
	"log"
	"time"
)

func main() {
	start := time.Now()
	fmt.Printf("string app ...\n\n")

	// connect db and migrate
	db.Connect()
	db.Migrate()
	gdb := db.GetGdb()

	itemSv := services.NewItemService(gdb)
	// vendorSv := services.NewVenderService(gdb)

	// upload items with excel(xlsx)
	itemsUpdateReport, err := itemSv.UpdateItemsFromExcel("zpath/excel/Items.xlsx")
	if err != nil {
		log.Printf("error while CreateBulk: %v", err)
	}
	// print result
	itemsUpdateReport.Print()

	text := "        ketofan   50  mg capsul    "
	log.Println(medName.RemoveExtraSapces(text))
	// // create new vendor
	// ven1, err := vendorSv.AddNew("الاهرام")
	// if err != nil {
	// 	log.Fatal("error while create vendor: ", err)
	// }

	// // upload vendor's offer
	// vendorOfferReport2, err := vendorSv.UpdateOffersFromExcel("zpath/excel/الاهرام.xlsx", ven1)
	// if err != nil {
	// 	log.Printf("error while updating vendor offer: %v", err)
	// }
	// // print result
	// vendorOfferReport2.Print()

	// // match all items
	// matchReport, err := vendorSv.AutoLinkItemsV3()
	// if err != nil {
	// 	log.Printf("error while match vendor's items offer: %v", err)
	// }
	// matchReport.Print()

	fmt.Printf("\napp take time: %v ", time.Since(start))
	fmt.Println("ending app ...")
}

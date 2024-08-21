package cmd

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/dataops"
	"github.com/scenery/mediax/routes"
)

var (
	migrate       = flag.Bool("migrate-db", false, "是否迁移数据库")
	importType    = flag.String("import", "", "导入数据来源: bangumi 或 douban")
	filePath      = flag.String("file", "", "导入文件的相对路径")
	downloadImage = flag.Bool("download-image", false, "可选，导入数据时是否下载图片")
	exportType    = flag.String("export", "", "导出数据: all, book, movie, tv, anime, game")
	limit         = flag.Int("limit", -1, "可选，指定导出的数量，默认导出所有数据")
	port          = flag.Int("port", config.HTTP_PORT, "指定启动端口")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Document: https://github.com/scenery/mediax/README.md")
		fmt.Println("\n** To import data, use the following command format:")
		fmt.Println("--import <douban|bangumi> --file <file.json> [--download-image]")
		fmt.Println("** To export data, use the following command format:")
		fmt.Println("--export <all|anime|movie|book|tv|game> [--limit <number>]")
		fmt.Println("\nAvailable parameters:")
		flag.PrintDefaults()
	}
}

func Execute() {
	flag.Parse()

	if *migrate {
		if *importType != "" || *exportType != "" || *filePath != "" || *downloadImage {
			fmt.Println("Error: --migrate-db cannot be used with other commands")
			os.Exit(1)
		}
		database.InitDB(true)
		return
	}

	if *importType != "" {
		if *filePath == "" {
			fmt.Println("Error: File path (relative path) is required for data import")
			flag.Usage()
			os.Exit(1)
		}
		database.InitDB(false)
		err := dataops.ImportFromJSON(*importType, *filePath, *downloadImage)
		if err != nil {
			fmt.Println("Error during import:", err)
			os.Exit(1)
		}
		return
	}
	if *downloadImage && *importType == "" {
		fmt.Println("Error: --download-image is only supported during import")
		flag.Usage()
		os.Exit(1)
	}

	if *exportType != "" {
		database.InitDB(false)
		err := dataops.ExportToJSON(*exportType, *limit)
		if err != nil {
			fmt.Println("Error during export:", err)
			os.Exit(1)
		}
		return
	}

	startServer(*port)
}

func startServer(port int) {
	database.InitDB(false)
	routes.Init()
	address := fmt.Sprintf(":%d", port)
	fmt.Print(`
                    _ _      __  __
 _ __ ___   ___  __| (_) __ _\ \/ /
| '_   _ \ / _ \/ _  | |/ _  |\  / 
| | | | | |  __/ (_| | | (_| |/  \ 
|_| |_| |_|\___|\__,_|_|\__,_/_/\_\

`)
	fmt.Printf("mediaX is up and running...\nPlease visit: http://localhost%s\n", address)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("mediaX failed to start: %v", err)
	}
}

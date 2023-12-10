package outbow

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"text/template"
)

func Main() int {
	slog.Debug("outbow", "test", true)

	run()
	return 0
}

func run() {
	myURL := &url.URL{
		Scheme:   "https",
		Host:     "gopro.com",
		Path:     "/en/us/shop/cameras/hero11-black/CHDHX-111-master.html",
		RawQuery: "yoReviewsPage=5",
	}

	tmpl, err := template.ParseFiles("gopro.tmpl")
	if err != nil {
		fmt.Println("Error reading template:", err)
		return
	}

	outputFile, err := os.Create("gopro.scpt")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	data := struct {
		MyURL string
	}{
		MyURL: myURL.String(),
	}
	err = tmpl.Execute(outputFile, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	fmt.Println("AppleScript generated and saved to gopro.scpt.")
}

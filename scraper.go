package main

import (
	"context"
	"fmt"

	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/emulation"

	//"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

//browser session data dir

const (
	browserDataDir = `~/.config/google-chrome/Default`
	source         = "https://justjoin.it/"
	minTimeS       = 5
	maxTimeS       = 10
	prefix         = "https://justjoin.it/job-offer/"
	offerSelector  = "a.offer-card\""
)

func getHTMLContent(chromeDpCtx context.Context, url string) (string, error) {
	var html string

	//chromdp run config
	err := chromedp.Run(
		chromeDpCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDeviceMetricsOverride(1280, 900, 1.0, false).Do(ctx)
		}),
		chromedp.Navigate(url),
		chromedp.Evaluate(`delete navigator.__proto__.webdriver`, nil),
		chromedp.Evaluate(`Object.defineProperty(navigator, "webdriver", { get: () => false })`, nil),
		chromedp.Sleep(time.Duration(rand.Intn(800)+300)*time.Millisecond),
		scrollToHell(),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	return html, err
}

func scrollToHell() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var prevHeight int64 = -1
		var currentHeight int64

		log.Println("Rozpoczynanie przewijania...")

		for i := 1; ; i++ {
			err := chromedp.Evaluate(`document.body.scrollHeight`, &currentHeight).Do(ctx)
			if err != nil {
				return fmt.Errorf("błąd podczas pobierania wysokości: %w", err)
			}

			if currentHeight == prevHeight {
				log.Printf("KONIEC: Wysokość strony nie zmieniła się (Wysokość: %d). Zatrzymanie.\n", currentHeight)
				break
			}

			prevHeight = currentHeight
			log.Printf("Iteracja %d: Przewijam do wysokości %d...\n", i, currentHeight)

			scrollScript := fmt.Sprintf(`window.scrollTo(0, %d);`, currentHeight)
			if err := chromedp.Evaluate(scrollScript, nil).Do(ctx); err != nil {
				return fmt.Errorf("błąd podczas przewijania: %w", err)
			}

			randomDelay := rand.Intn(maxTimeS-minTimeS) + minTimeS
			time.Sleep(time.Duration(randomDelay))
		}

		return nil
	}
}

func getUrlsFromContent(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Printf("goquery parse error: %v", err)
		return nil, err
	}

	var urls []string

	doc.Find(offerSelector).Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			urls = append(urls, prefix+href)
		}
	})

	return urls, nil
}

func CollectJustJoinIt(ctx context.Context) []string {
	var urls []string

	//chromdp config
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("/usr/bin/google-chrome"),
		chromedp.UserDataDir(browserDataDir),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"),
		//chromedp.Flag("proxy-server", proxyList[rand.Intn(len(proxyList))]),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-site-isolation-trials", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	chromeDpCtx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	html, err := getHTMLContent(chromeDpCtx, source)
	if err != nil {
		log.Printf("Error getting HTML content", err)
	}

	urls, err = getUrlsFromContent(html)
	if err != nil {
		log.Printf("Error getting urls", err)
	} else {
		log.Printf("Scraping succesfull, collected %v urls", len(urls))
	}

	return urls
}

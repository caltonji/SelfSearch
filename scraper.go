package main

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/mauidude/go-readability"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Result struct {
	Title   string
	Image   string
	Content string
}

func scrape(queryUrl string) (*Result, error) {
	// get the html text of the site
	res, err := http.Get(queryUrl)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("Non 200 response when scraping site")
	}
	htmlBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()
	html := string(htmlBytes)

	// Get the metadata from the header
	fullReader := strings.NewReader(html)

	fullDoc, err := goquery.NewDocumentFromReader(fullReader)
	if err != nil {
		return nil, err
	}

	pageTitle := ""
	fullDoc.Find("meta[property='og:title']").Each(func(i int, s *goquery.Selection) {
		pageTitle, _ = s.Attr("content")
	})

	imageUrl := ""
	fullDoc.Find("meta[property='og:image']").Each(func(i int, s *goquery.Selection) {
		imageUrlString, _ := s.Attr("content")
		imageUrlRaw, _ := url.Parse(imageUrlString)
		imageUrl = path.Join(imageUrlRaw.Host, imageUrlRaw.Path)
	})

	// Get the abridged content
	readabilityDoc, err := readability.NewDocument(html)
	if err != nil {
		return nil, err
	}

	abridgedHtml := readabilityDoc.Content()
	abridgedReader := strings.NewReader(abridgedHtml)
	abridgedDoc, err := goquery.NewDocumentFromReader(abridgedReader)
	if err != nil {
		return nil, err
	}
	abridgedText := abridgedDoc.Text()

	return &Result{
		Title:   pageTitle,
		Image:   imageUrl,
		Content: abridgedText,
	}, nil
}

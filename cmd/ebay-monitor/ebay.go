package main

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type Format string

const (
	Auction  Format = "auction"
	BuyItNow Format = "buy-it-now"
)

type Listing struct {
	Url            string `json:"url"`
	ImageUrl       string `json:"imageUrl"`
	EbayItemNumber string `json:"ebayItemNumber"`

	SellerName               string  `json:"sellerName"`
	SellerStars              int     `json:"sellerStars"`
	SellerFeedbackPercentage float32 `json:"sellerFeedbackPercentage"`

	Format       Format  `json:"format"`
	Location     string  `json:"location"`
	Title        string  `json:"title"`
	Condition    string  `json:"condition"`
	Price        float32 `json:"price"`
	Currency     string
	Postage      int    `json:"postage"`
	CanMakeOffer bool   `json:"canMakeOffer"`
	Returns      string `json:"returns"`
}

func GetPrice(price string) (float32, error) {
	// some countries use commas instead of dots so replace commas with dots
	if strings.Index(price, ",") != -1 {
		price = strings.ReplaceAll(price, ".", "")
		price = strings.ReplaceAll(price, ",", ".")
	}
	foundStartingIndex := false
	var startingIndex, endingIndex int
	for i, rune := range price {
		if unicode.IsDigit(rune) || rune == '.' {
			if !foundStartingIndex {
				startingIndex = i
				foundStartingIndex = true
			}
			continue
		}
		if foundStartingIndex {
			endingIndex = i - 1
			break
		}
	}

	if endingIndex == 0 {
		endingIndex = len(price)
	}

	float, err := strconv.ParseFloat(price[startingIndex:endingIndex], 32)
	if err != nil {
		return 0, errors.Wrap(err, "could not convert price to float")
	}
	return float32(float), nil
}

func GetListing(url string, currency string, doc *goquery.Document) (*Listing, error) {
	listing := &Listing{}

	listing.Url = url

	// Get listing image
	imageUrl, exists := doc.Find("div.ux-image-carousel-item").First().Find("img").Attr("src")
	if !exists {
		return nil, errors.New("Could not find image element")
	}
	listing.ImageUrl = imageUrl

	// Get listing seller
	listing.SellerName = doc.Find("div.x-sellercard-atf__info__about-seller").Find("a").Find("span.ux-textspans").Text()

	// Get listing price
	price, err := GetPrice(doc.Find("div.x-price-primary").Find("span.ux-textspans").Text())
	if err != nil {
		return nil, errors.Wrap(err, "could not determine price")
	}
	listing.Price = price

	// Get listing format (BIN vs auction)
	format := BuyItNow
	if doc.Find("li").Find("div.vim x-bid-action").Length() > 0 {
		format = Auction
	}
	listing.Format = format

	// Get listing title
	listing.Title = doc.Find("h1.x-item-title__mainTitle").Find("span.ux-textspans").Text()

	// TODO: re-implement this later
	// listing.CanMakeOffer = false
	// if len(doc.Find("a#boBtn_btn").Nodes) > 0 {
	// 	listing.CanMakeOffer = true
	// }

	return listing, nil
}

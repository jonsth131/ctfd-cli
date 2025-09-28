package api

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func extractNonce(htmlBody string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return "", err
	}

	var nonce string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var id, name, value string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "id":
					id = attr.Val
				case "name":
					name = attr.Val
				case "value":
					value = attr.Val
				}
			}
			if id == "nonce" && name == "nonce" {
				nonce = value
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	if nonce == "" {
		return "", fmt.Errorf("nonce not found")
	}
	return nonce, nil
}

func extractTitle(htmlBody string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return "", err
	}

	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					title = strings.TrimSpace(c.Data)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	if title == "" {
		return "", fmt.Errorf("title not found")
	}
	return title, nil
}

func extractCSRFToken(htmlBody string) (string, error) {
	regex := regexp.MustCompile(`'csrfNonce': "([a-f0-9]*)",`)
	match := regex.FindStringSubmatch(htmlBody)

	if len(match) != 2 {
		return "", fmt.Errorf("csrf token not found")
	}

	nonce := match[1]

	return nonce, nil
}

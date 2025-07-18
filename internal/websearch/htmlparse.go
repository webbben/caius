package websearch

import (
	"strings"

	"golang.org/x/net/html"
)

func RemoveAllHTML(s string) string {
	var sb strings.Builder
	inTag := false

	for _, c := range s {
		if inTag {
			// continue until we find the end of the tag
			if c == '>' {
				inTag = false
			}
			continue
		}
		if c == '<' {
			// we have now entered a tag
			inTag = true
			continue
		}
		sb.WriteRune(c)
	}

	return sb.String()
}

func findNode(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res := findNode(c, tagName)
		if res != nil {
			return res
		}
	}

	return nil
}

func extractTextFromNode(n *html.Node) string {
	var buf strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(" ")
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(buf.String())
}

func extractBodyText(s string) (string, error) {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return "", err
	}

	// get the body
	bodyNode := findNode(doc, "body")
	if bodyNode == nil {
		return "", nil // not found
	}

	cleanHTML(bodyNode)

	return extractTextFromNode(bodyNode), nil
}

func removeHTMLTag(n *html.Node, tagName string) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		if c.Type == html.ElementNode && c.Data == tagName {
			n.RemoveChild(c)
		} else {
			removeHTMLTag(c, tagName)
		}
		c = next
	}
}

func cleanHTML(n *html.Node) {
	removeTags := []string{"style", "script", "meta", "link", "svg", "header", "iframe", "button", "h1", "h2"}

	for _, tag := range removeTags {
		removeHTMLTag(n, tag)
	}
}

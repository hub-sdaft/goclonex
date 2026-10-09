package goclonex

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	pkgUrl "net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/hub-sdaft/goclonex/pkg/crawler"
	"github.com/hub-sdaft/goclonex/pkg/html"
	"github.com/hub-sdaft/goclonex/pkg/parser"
	"github.com/hub-sdaft/goclonex/pkg/project"
)

type CloneOptions struct {
	IgnoreJS,
	IgnoreCSS,
	IgnoreImages bool

	Cookies   []string
	Proxy     string
	UserAgent string
}

func setupCookieJar(urlString string, cookieStrings []string) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}

	if len(cookieStrings) == 0 {
		return jar, nil
	}

	cookies := make([]*http.Cookie, 0, len(cookieStrings))
	for _, cookieString := range cookieStrings {
		fields := strings.FieldsSeq(cookieString)
		for field := range fields {
			var k, v string
			if i := strings.IndexByte(field, '='); i >= 0 {
				k, v = field[:i], strings.TrimRight(field[i+1:], ";")
			} else {
				return nil, fmt.Errorf("no = in cookie %q", cookieString)
			}
			cookies = append(cookies, &http.Cookie{Name: k, Value: v})
		}
	}

	url, err := pkgUrl.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", urlString, err)
	}
	jar.SetCookies(&pkgUrl.URL{Scheme: url.Scheme, User: url.User, Host: url.Host}, cookies)

	return jar, nil
}

func NewProject(name, url string) project.Project {
	return project.NewProject(name, url)
}

func Clone(proj *project.Project, opt CloneOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	url := proj.Url()

	// 1. Validate url
	isValid, isValidDomain := parser.ValidateURL(url), parser.ValidateDomain(url)
	if !isValid && !isValidDomain {
		return fmt.Errorf("%q is not valid", url)
	}

	// 2. Setup cookies
	cookieJar, err := setupCookieJar(url, opt.Cookies)
	if err != nil {
		return fmt.Errorf("error creating cookie jar %s: %w", opt.Cookies, err)
	}

	// 3. Crawl
	crawlOpt := crawler.CrawlOptions{
		IgnoreJS:     opt.IgnoreJS,
		IgnoreCSS:    opt.IgnoreCSS,
		IgnoreImages: opt.IgnoreImages,
		CookieJar:    cookieJar,
		Proxy:        opt.Proxy,
		UserAgent:    opt.UserAgent,
	}
	if err := crawler.CrawlProject(ctx, proj, crawlOpt); err != nil {
		return fmt.Errorf("error crawling %q: %w", url, err)
	}

	// 4. Restructure links
	if err := html.ProjectLinkRestructure(proj); err != nil {
		return fmt.Errorf("error restructuring %q: %w", url, err)
	}

	// 5. Format
	html.ProjectFormatHTML(proj)

	return nil
}

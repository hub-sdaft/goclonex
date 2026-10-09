package crawler

import (
	"context"
	"net/http/cookiejar"

	"github.com/hub-sdaft/goclonex/pkg/project"
)

// Crawl asks the necessary crawlers for collecting links for building the web page
func Crawl(ctx context.Context, site string, projectPath string, cookieJar *cookiejar.Jar, proxyString string, userAgent string) error {
	// searches for css, js, and images within a given link
	return Collector(ctx, site, projectPath, cookieJar, proxyString, userAgent)
}

type CrawlOptions struct {
	IgnoreJS,
	IgnoreCSS,
	IgnoreImages bool
	
	CookieJar *cookiejar.Jar
	Proxy     string
	UserAgent string
}

func CrawlProject(ctx context.Context, p *project.Project, opt CrawlOptions) error {
	return ProjectCollector(ctx, p, opt)
}

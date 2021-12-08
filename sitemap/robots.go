package sitemap

import (
	"fmt"
	"strings"
)

type RobotsBuilder struct {
	userAgents []*userAgentBuilder
}

type userAgentBuilder struct {
	name      string
	disallows []string
	allows    []string
	sitemaps  []string
}

const (
	// https://www.keycdn.com/blog/web-crawlers
	AllAgents     = "*"
	GoogleAgent   = "Googlebot"
	BingAgent     = "Bingbot"
	YahooAgent    = "Slurp"
	DuckDuckAgent = "DuckDuckBot"
	BaiduAgent    = "Baiduspider"
	YandexAgent   = "YandexBot"
	SogouAgent    = "Sogou web spider/4.0"
	ExaleadAgent  = "Mozilla/5.0 (compatible; Konqueror/3.5; Linux) KHTML/3.5.5 (like Gecko) (Exabot-Thumbnails)"
	FacebookAgent = "facebot"
	AlexaAgent    = "ia_archiver"
)

func Robots() *RobotsBuilder {
	return &RobotsBuilder{}
}

func (r *RobotsBuilder) Agent(name string) *userAgentBuilder {
	agent := &userAgentBuilder{
		name: name,
	}
	r.userAgents = append(r.userAgents, agent)
	return agent
}

func (r *RobotsBuilder) ToTxt() string {
	b := strings.Builder{}
	for _, agent := range r.userAgents {
		b.WriteString(fmt.Sprintf("User-agent: %s\n", agent.name))
		for _, disallow := range agent.disallows {
			b.WriteString(fmt.Sprintf("Disallow: %s\n", disallow))
		}
		for _, allow := range agent.allows {
			b.WriteString(fmt.Sprintf("Allow: %s\n", allow))
		}
		for _, sitemap := range agent.sitemaps {
			b.WriteString(fmt.Sprintf("Sitemap: %s\n", sitemap))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (u *userAgentBuilder) AddSitemapUrl(sitemaps ...string) *userAgentBuilder {
	u.sitemaps = append(u.sitemaps, sitemaps...)
	return u
}

func (u *userAgentBuilder) Allow(allows ...string) *userAgentBuilder {
	u.allows = append(u.allows, allows...)
	return u
}

func (u *userAgentBuilder) Disallow(disallows ...string) *userAgentBuilder {
	u.disallows = append(u.disallows, disallows...)
	return u
}

package i18n

import (
	"fmt"
	"strings"
)

type Dict map[string]string

var translations = map[string]Dict{
	"en": {
		"skip_to_content":    "Skip to content",
		"toggle_theme":       "Toggle color theme",
		"built_with":         "Built with Alster",
		"main_nav":           "Main",
		"social_nav":         "Social",
		"sections_nav":       "Sections",
		"projects":           "Projects",
		"projects_desc":      "Selected tools, themes & experiments",
		"projects_lead":      "Selected work, tools, and technical experiments.",
		"writing":            "Writing",
		"writing_desc":       "Notes, thoughts & technical write-ups",
		"writing_lead":       "Notes, thoughts, and technical write-ups.",
		"about":              "About Me",
		"about_desc":         "Background, philosophy & colophon",
		"recently_shipped":   "Recently shipped",
		"all_projects":       "All projects →",
		"latest_writing":     "Latest writing",
		"all_writing":        "All writing →",
		"latest_activity":    "Latest",
		"feed_type_project":  "Project",
		"feed_type_post":     "Writing",
		"filter_all":         "All",
		"filter_aria":        "Filter projects by tag",
		"no_projects":        "No projects yet. Add a Markdown file inside a content folder.",
		"no_projects_found":  "No projects found with this tag.",
		"no_posts":           "No posts yet.",
		"back_to_projects":   "← Projects",
		"back_to_writing":    "← Writing",
		"live_project":       "Live project ↗",
		"source_code":        "Source code ↗",
		"status_wip":         "In progress",
		"connect":            "Connect",
		"sidebar_label":      "Sidebar",
		"default_bio":        "I'm %s, a software engineer and designer. Focused on simple tools, idiomatic Go, and calm interfaces.",
		"default_about_html": "<p>%s is a creator and engineer building minimalist software and design systems.</p>",
	},
	"sv": {
		"skip_to_content":    "Hoppa till innehåll",
		"toggle_theme":       "Växla färgtema",
		"built_with":         "Byggd med Alster",
		"main_nav":           "Huvudmeny",
		"social_nav":         "Socialt",
		"sections_nav":       "Sektioner",
		"projects":           "Projekt",
		"projects_desc":      "Utvalda verktyg, teman & experiment",
		"projects_lead":      "Utvalda arbeten, verktyg och tekniska experiment.",
		"writing":            "Skrivande",
		"writing_desc":       "Anteckningar, tankar & tekniska texter",
		"writing_lead":       "Anteckningar, tankar och tekniska texter.",
		"about":              "Om mig",
		"about_desc":         "Bakgrund, filosofi & kolofon",
		"recently_shipped":   "Senast skapat",
		"all_projects":       "Alla projekt →",
		"latest_writing":     "Senast skrivet",
		"all_writing":        "Alla texter →",
		"latest_activity":    "Senaste",
		"feed_type_project":  "Projekt",
		"feed_type_post":     "Text",
		"filter_all":         "Alla",
		"filter_aria":        "Filtrera projekt efter tagg",
		"no_projects":        "Inga projekt än. Lägg till en Markdown-fil i en innehållsmapp.",
		"no_projects_found":  "Inga projekt hittades med denna tagg.",
		"no_posts":           "Inga inlägg än.",
		"back_to_projects":   "← Projekt",
		"back_to_writing":    "← Skrivande",
		"live_project":       "Webbplats ↗",
		"source_code":        "Källkod ↗",
		"status_wip":         "Pågående",
		"connect":            "Kontakt",
		"sidebar_label":      "Sidopanel",
		"default_bio":        "Jag är %s, mjukvaruingenjör och designer. Fokuserad på enkla verktyg, idiomatisk Go och lugna gränssnitt.",
		"default_about_html": "<p>%s är en skapare och ingenjör som bygger minimalistisk mjukvara och designsystem.</p>",
	},
}

// T returns the translated string for lang and key, falling back to English or key if missing.
func T(lang, key string, args ...any) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = "en"
	}
	dict, ok := translations[lang]
	if !ok {
		dict = translations["en"]
	}
	val, ok := dict[key]
	if !ok {
		if enDict, ok := translations["en"]; ok {
			val = enDict[key]
		}
	}
	if val == "" {
		val = key
	}
	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// IsSectionMatch determines if a given navigation label matches the active section in English or Swedish.
func IsSectionMatch(section, label string) bool {
	if strings.EqualFold(section, label) {
		return true
	}
	sec := strings.ToLower(strings.TrimSpace(section))
	lbl := strings.ToLower(strings.TrimSpace(label))
	switch sec {
	case "projects":
		return lbl == "projekt" || lbl == "projects"
	case "blog":
		return lbl == "blog" || lbl == "blogg" || lbl == "writing" || lbl == "skrivande" || lbl == "texter"
	case "about":
		return lbl == "about" || lbl == "about me" || lbl == "om" || lbl == "om mig"
	}
	return false
}

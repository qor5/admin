package testflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedMatchOne(t *testing.T) {
	text := `
	<v-navigation-drawer v-model='vars.presetsRightDrawer' :location='"right"' :temporary='true' :width='"600"' :height='"100%"' class='v-navigation-drawer--temporary'>
<global-events @keyup.esc='vars.presetsRightDrawer = false'></global-events>

<go-plaid-portal :visible='true' :form='form' :locals='locals' portal-name='presets_RightDrawerContentPortalName'>
<go-plaid-scope v-slot='{ form }'>
<v-layout>
<v-app-bar color='white' :elevation='0'>
<v-app-bar-title class='pl-2'>
<div class='d-flex'>WithPublishProduct 7_2024-05-23-v01</div>
</v-app-bar-title>

<v-btn :icon='"mdi-close"' @click.stop='vars.presetsRightDrawer = false'></v-btn>
</v-app-bar></v-navigation-drawer>`
	m, err := NamedMatchOne(`<v-navigation-drawer v-model='vars.presetsRightDrawer'[\s\S]+?(<v-app-bar-title[^>]+>\s*<div[^>]+>(?P<title>.+?)\s*<\/div>\s*<\/v-app-bar-title>|<v-toolbar-title[^>]+>(?P<title>.+?)<\/v-toolbar-title>)[\s\S]+?<\/v-navigation-drawer>`, text)
	require.NoError(t, err)
	assert.Equal(t, "WithPublishProduct 7_2024-05-23-v01", m["title"])
}

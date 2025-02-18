import install from '@twind/with-web-components'
import { defineConfig } from '@twind/core'
import presetAutoprefix from '@twind/preset-autoprefix'
import presetTailwind from '@twind/preset-tailwind'
import Alpine from 'alpinejs'

declare global {
  interface Window {
    TwindScope: any
    twindConfig: any
    Alpine: any
  }
}

type StyleType = 'url' | 'inlineStyle'
type ScriptType = 'url' | 'inlineScript'
;(function (win) {
  const stylesList: Array<{ str: string; type: StyleType }> = []
  const scriptsList: Array<{ str: string; type: ScriptType }> = []

  initStyleAndScript(win.TwindScope?.style || [], win.TwindScope?.script || [])

  function initStyleAndScript(styleList: string[], scriptList: string[]) {
    styleList.forEach((str) => {
      let type = 'inlineStyle'
      if (/^https?:\/\//.test(str)) type = 'url'

      stylesList.push({
        type: type as 'url' | 'inlineStyle',
        str,
      })
    })

    scriptList.forEach((str) => {
      let type = 'inlineScript'
      if (/^https?:\/\//.test(str)) type = 'url'

      scriptsList.push({
        type: type as 'url' | 'inlineScript',
        str,
      })
    })
  }

  const withTwind = install(
    defineConfig({
      presets: [presetAutoprefix(), presetTailwind()],
      ...(win.TwindScope?.config || {}),
    })
  )
  class TwindScope extends withTwind(HTMLElement) {
    constructor() {
      super()
      this.attachShadow({ mode: 'open' })

      this.integrateStyleAndScript()

      if (this.shadowRoot) {
        this.shadowRoot.innerHTML = this.innerHTML
        this.innerHTML = ''
        Alpine.initTree(this.shadowRoot.firstElementChild as HTMLElement)
      }
    }

    disconnectedCallback(): void {
      this.shadowRoot &&
        Alpine.destroyTree(this.shadowRoot.firstElementChild as HTMLElement)
    }

    integrateStyleAndScript() {
      if (stylesList.length > 0) {
        stylesList.forEach((style) => {
          if (!this.shadowRoot) return

          if (style.type === 'inlineStyle') {
            const sheet = new CSSStyleSheet()
            sheet.replaceSync(style.str)
            this.shadowRoot.adoptedStyleSheets = [sheet]
          }

          if (style.type === 'url') {
            const link = document.createElement('link')
            link.rel = 'stylesheet'
            link.href = style.str
            this.shadowRoot.appendChild(link)
          }
        })
      }

      if (scriptsList.length > 0) {
        scriptsList.forEach((script) => {
          if (!this.shadowRoot) return

          if (script.type === 'inlineScript') {
            const scriptElement = document.createElement('script')
            scriptElement.textContent = script.str
            this.shadowRoot.appendChild(scriptElement)
          }

          if (script.type === 'url') {
            const scriptElement = document.createElement('script')
            scriptElement.src = script.str
            this.shadowRoot.appendChild(scriptElement)
          }
        })
      }
    }
  }

  customElements.define('twind-scope', TwindScope)
})(window)

import install from '@twind/with-web-components'
import config from '../twind.config.ts'
;(function () {
  const withTwind = install(config)

  class TwindElement extends withTwind(HTMLElement) {
    constructor() {
      super()
      this.attachShadow({ mode: 'open' })
      if (this.shadowRoot) {
        this.shadowRoot.innerHTML = this.innerHTML
        this.innerHTML = ''
      }
    }
  }

  customElements.define('twind-scope', TwindElement)
})()

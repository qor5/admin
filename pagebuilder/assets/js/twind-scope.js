function xe(e) {
  return [
    ...e.v,
    (e.i ? "!" : "") + e.n
  ].join(":");
}
function Xe(e, t = ",") {
  return e.map(xe).join(t);
}
let Le = typeof CSS < "u" && CSS.escape || // Simplified: escaping only special characters
// Needed for NodeJS and Edge <79 (https://caniuse.com/mdn-api_css_escape)
((e) => e.replace(/[!"'`*+.,;:\\/<=>?@#$%&^|~()[\]{}]/g, "\\$&").replace(/^\d/, "\\3$& "));
function te(e) {
  for (var t = 9, r = e.length; r--; ) t = Math.imul(t ^ e.charCodeAt(r), 1597334677);
  return "#" + ((t ^ t >>> 9) >>> 0).toString(36);
}
function ve(e, t = "@media ") {
  return t + w(e).map((r) => (typeof r == "string" && (r = {
    min: r
  }), r.raw || Object.keys(r).map((o) => `(${o}-width:${r[o]})`).join(" and "))).join(",");
}
function w(e = []) {
  return Array.isArray(e) ? e : e == null ? [] : [
    e
  ];
}
function Te(e) {
  return e;
}
function ke() {
}
let T = {
  /**
  * 1. `default` (public)
  */
  d: (
    /* efaults */
    0
  ),
  /* Shifts.layer */
  /**
  * 2. `base` (public) — for things like reset rules or default styles applied to plain HTML elements.
  */
  b: (
    /* ase */
    134217728
  ),
  /* Shifts.layer */
  /**
  * 3. `components` (public, used by `style()`) — is for class-based styles that you want to be able to override with utilities.
  */
  c: (
    /* omponents */
    268435456
  ),
  /* Shifts.layer */
  // reserved for style():
  // - props: 0b011
  // - when: 0b100
  /**
  * 6. `aliases` (public, used by `apply()`) — `~(...)`
  */
  a: (
    /* liases */
    671088640
  ),
  /* Shifts.layer */
  /**
  * 6. `utilities` (public) — for small, single-purpose classes
  */
  u: (
    /* tilities */
    805306368
  ),
  /* Shifts.layer */
  /**
  * 7. `overrides` (public, used by `css()`)
  */
  o: (
    /* verrides */
    939524096
  )
};
function Ue(e) {
  var t;
  return ((t = e.match(/[-=:;]/g)) == null ? void 0 : t.length) || 0;
}
function ue(e) {
  return Math.min(/(?:^|width[^\d]+)(\d+(?:.\d+)?)(p)?/.test(e) ? Math.max(0, 29.63 * (+RegExp.$1 / (RegExp.$2 ? 15 : 1)) ** 0.137 - 43) : 0, 15) << 22 | /* Shifts.responsive */
  Math.min(Ue(e), 15) << 18;
}
let Ze = [
  /* fi */
  "rst-c",
  /* hild: 0 */
  /* la */
  "st-ch",
  /* ild: 1 */
  // even and odd use: nth-child
  /* nt */
  "h-chi",
  /* ld: 2 */
  /* an */
  "y-lin",
  /* k: 3 */
  /* li */
  "nk",
  /* : 4 */
  /* vi */
  "sited",
  /* : 5 */
  /* ch */
  "ecked",
  /* : 6 */
  /* em */
  "pty",
  /* : 7 */
  /* re */
  "ad-on",
  /* ly: 8 */
  /* fo */
  "cus-w",
  /* ithin : 9 */
  /* ho */
  "ver",
  /* : 10 */
  /* fo */
  "cus",
  /* : 11 */
  /* fo */
  "cus-v",
  /* isible : 12 */
  /* ac */
  "tive",
  /* : 13 */
  /* di */
  "sable",
  /* d : 14 */
  /* op */
  "tiona",
  /* l: 15 */
  /* re */
  "quire"
];
function Se({ n: e, i: t, v: r = [] }, o, n, l) {
  e && (e = xe({
    n: e,
    i: t,
    v: r
  })), l = [
    ...w(l)
  ];
  for (let a of r) {
    let i = o.theme("screens", a);
    for (let f of w(i && ve(i) || o.v(a))) {
      var s;
      l.push(f), n |= i ? 67108864 | /* Shifts.screens */
      ue(f) : a == "dark" ? 1073741824 : (
        /* Shifts.darkMode */
        f[0] == "@" ? ue(f) : (s = f, // use first found pseudo-class
        1 << ~(/:([a-z-]+)/.test(s) && ~Ze.indexOf(RegExp.$1.slice(2, 7)) || -18))
      );
    }
  }
  return {
    n: e,
    p: n,
    r: l,
    i: t
  };
}
let Ie = /* @__PURE__ */ new Map();
function ge(e) {
  if (e.d) {
    let t = [], r = ae(
      // merge all conditions into a selector string
      e.r.reduce((o, n) => n[0] == "@" ? (t.push(n), o) : (
        // Go over the selector and replace the matching multiple selectors if any
        n ? ae(o, (l) => ae(
          n,
          // If the current condition has a nested selector replace it
          (s) => {
            let a = /(:merge\(.+?\))(:[a-z-]+|\\[.+])/.exec(s);
            if (a) {
              let i = l.indexOf(a[1]);
              return ~i ? (
                // [':merge(.group):hover .rule', ':merge(.group):focus &'] -> ':merge(.group):focus:hover .rule'
                // ':merge(.group)' + ':focus' + ':hover .rule'
                l.slice(0, i) + a[0] + l.slice(i + a[1].length)
              ) : (
                // [':merge(.peer):focus~&', ':merge(.group):hover &'] -> ':merge(.peer):focus~:merge(.group):hover &'
                le(l, s)
              );
            }
            return le(s, l);
          }
        )) : o
      ), "&"),
      // replace '&' with rule name or an empty string
      (o) => le(o, e.n ? "." + Le(e.n) : "")
    );
    return r && t.push(r.replace(/:merge\((.+?)\)/g, "$1")), t.reduceRight((o, n) => n + "{" + o + "}", e.d);
  }
}
function ae(e, t) {
  return e.replace(/ *((?:\(.+?\)|\[.+?\]|[^,])+) *(,|$)/g, (r, o, n) => t(o) + n);
}
function le(e, t) {
  return e.replace(/&/g, t);
}
let ze = new Intl.Collator("en", {
  numeric: !0
});
function Be(e, t) {
  for (var r = 0, o = e.length; r < o; ) {
    let n = o + r >> 1;
    0 >= He(e[n], t) ? r = n + 1 : o = n;
  }
  return o;
}
function He(e, t) {
  let r = e.p & T.o;
  return r == (t.p & T.o) && (r == T.b || r == T.o) ? 0 : e.p - t.p || e.o - t.o || ze.compare(Ae(e.n), Ae(t.n)) || ze.compare(Fe(e.n), Fe(t.n));
}
function Ae(e) {
  return (e || "").split(/:/).pop().split("/").pop() || "\0";
}
function Fe(e) {
  return (e || "").replace(/\W/g, (t) => String.fromCharCode(127 + t.charCodeAt(0))) + "\0";
}
function se(e, t) {
  return Math.round(parseInt(e, 16) * t);
}
function U(e, t = {}) {
  if (typeof e == "function") return e(t);
  let { opacityValue: r = "1", opacityVariable: o } = t, n = o ? `var(${o})` : r;
  if (e.includes("<alpha-value>")) return e.replace("<alpha-value>", n);
  if (e[0] == "#" && (e.length == 4 || e.length == 7)) {
    let l = (e.length - 1) / 3, s = [
      17,
      1,
      0.062272
    ][l - 1];
    return `rgba(${[
      se(e.substr(1, l), s),
      se(e.substr(1 + l, l), s),
      se(e.substr(1 + 2 * l, l), s),
      n
    ]})`;
  }
  return n == "1" ? e : n == "0" ? "#0000" : (
    // convert rgb and hsl to alpha variant
    e.replace(/^(rgb|hsl)(\([^)]+)\)$/, `$1a$2,${n})`)
  );
}
function Ne(e, t, r, o, n = []) {
  return function l(s, { n: a, p: i, r: f = [], i: p }, u) {
    let g = [], y = "", x = 0, v = 0;
    for (let b in s || {}) {
      var $, B;
      let k = s[b];
      if (b[0] == "@") {
        if (!k) continue;
        if (b[1] == "a") {
          g.push(...$e(a, i, oe("" + k), u, i, f, p, !0));
          continue;
        }
        if (b[1] == "l") {
          for (let z of w(k)) g.push(...l(z, {
            n: a,
            p: ($ = T[b[7]], // Set layer (first reset, than set)
            i & -939524097 | $),
            r: b[7] == "d" ? [] : f,
            i: p
          }, u));
          continue;
        }
        if (b[1] == "i") {
          g.push(...w(k).map((z) => ({
            // before all layers
            p: -1,
            o: 0,
            r: [],
            d: b + " " + z
          })));
          continue;
        }
        if (b[1] == "k") {
          g.push({
            p: T.d,
            o: 0,
            r: [
              b
            ],
            d: l(k, {
              p: T.d
            }, u).map(ge).join("")
          });
          continue;
        }
        if (b[1] == "f") {
          g.push(...w(k).map((z) => ({
            p: T.d,
            o: 0,
            r: [
              b
            ],
            d: l(z, {
              p: T.d
            }, u).map(ge).join("")
          })));
          continue;
        }
      }
      if (typeof k != "object" || Array.isArray(k))
        b == "label" && k ? a = k + te(JSON.stringify([
          i,
          p,
          s
        ])) : (k || k === 0) && (b = b.replace(/[A-Z]/g, (z) => "-" + z.toLowerCase()), v += 1, x = Math.max(x, (B = b)[0] == "-" ? 0 : Ue(B) + (/^(?:(border-(?!w|c|sty)|[tlbr].{2,4}m?$|c.{7,8}$)|([fl].{5}l|g.{8}$|pl))/.test(B) ? +!!RegExp.$1 || /* +1 */
        -!!RegExp.$2 : (
          /* -1 */
          0
        )) + 1), y += (y ? ";" : "") + w(k).map((z) => u.s(
          b,
          // support theme(...) function in values
          // calc(100vh - theme('spacing.12'))
          Ce("" + z, u.theme) + (p ? " !important" : "")
        )).join(";"));
      else if (b[0] == "@" || b.includes("&")) {
        let z = i;
        b[0] == "@" && (b = b.replace(/\bscreen\(([^)]+)\)/g, (ne, q) => {
          let M = u.theme("screens", q);
          return M ? (z |= 67108864, /* Shifts.screens */
          ve(M, "")) : ne;
        }), z |= ue(b)), g.push(...l(k, {
          n: a,
          p: z,
          r: [
            ...f,
            b
          ],
          i: p
        }, u));
      } else
        g.push(...l(k, {
          p: i,
          r: [
            ...f,
            b
          ]
        }, u));
    }
    return (
      // PERF: prevent unshift using `rules = [{}]` above and then `rules[0] = {...}`
      g.unshift({
        n: a,
        p: i,
        o: (
          // number of declarations (descending)
          Math.max(0, 15 - v) + // greatest precedence of properties
          // if there is no property precedence this is most likely a custom property only declaration
          // these have the highest precedence
          1.5 * Math.min(x || 15, 15)
        ),
        r: f,
        // stringified declarations
        d: y
      }), g.sort(He)
    );
  }(e, Se(t, r, o, n), r);
}
function Ce(e, t) {
  return e.replace(/theme\((["'`])?(.+?)\1(?:\s*,\s*(["'`])?(.+?)\3)?\)/g, (r, o, n, l, s = "") => {
    let a = t(n, s);
    return typeof a == "function" && /color|fill|stroke/i.test(n) ? U(a) : "" + w(a).filter((i) => Object(i) !== i);
  });
}
function Pe(e, t) {
  let r, o = [];
  for (let n of e)
    n.d && n.n ? (r == null ? void 0 : r.p) == n.p && "" + r.r == "" + n.r ? (r.c = [
      r.c,
      n.c
    ].filter(Boolean).join(" "), r.d = r.d + ";" + n.d) : o.push(r = {
      ...n,
      n: n.n && t
    }) : o.push({
      ...n,
      n: n.n && t
    });
  return o;
}
function re(e, t, r = T.u, o, n) {
  let l = [];
  for (let s of e) for (let a of function(i, f, p, u, g) {
    i = {
      ...i,
      i: i.i || g
    };
    let y = function(x, v) {
      let $ = Ie.get(x.n);
      return $ ? $(x, v) : v.r(x.n, x.v[0] == "dark");
    }(i, f);
    return y ? (
      // a list of class names
      typeof y == "string" ? ({ r: u, p } = Se(i, f, p, u), Pe(re(oe(y), f, p, u, i.i), i.n)) : Array.isArray(y) ? y.map((x) => {
        var v, $;
        return {
          o: 0,
          ...x,
          r: [
            ...w(u),
            ...w(x.r)
          ],
          p: (v = p, $ = x.p ?? p, v & -939524097 | $)
        };
      }) : Ne(y, i, f, p, u)
    ) : (
      // propagate className as is
      [
        {
          c: xe(i),
          p: 0,
          o: 0,
          r: []
        }
      ]
    );
  }(s, t, r, o, n)) l.splice(Be(l, a), 0, a);
  return l;
}
function $e(e, t, r, o, n, l, s, a) {
  return Pe((a ? r.flatMap((i) => re([
    i
  ], o, n, l, s)) : re(r, o, n, l, s)).map((i) => (
    // do not move defaults
    // move only rules with a name unless they are in the base layer
    i.p & T.o && (i.n || t == T.b) ? {
      ...i,
      p: i.p & -939524097 | t,
      o: 0
    } : i
  )), e);
}
function Qe(e, t, r, o) {
  var n;
  return n = (l, s) => {
    let { n: a, p: i, r: f, i: p } = Se(l, s, t);
    return r && $e(a, t, r, s, i, f, p, o);
  }, Ie.set(e, n), e;
}
function ce(e, t, r) {
  if (e[e.length - 1] != "(") {
    let o = [], n = !1, l = !1, s = "";
    for (let a of e) if (!(a == "(" || /[~@]$/.test(a))) {
      if (a[0] == "!" && (a = a.slice(1), n = !n), a.endsWith(":")) {
        o[a == "dark:" ? "unshift" : "push"](a.slice(0, -1));
        continue;
      }
      a[0] == "-" && (a = a.slice(1), l = !l), a.endsWith("-") && (a = a.slice(0, -1)), a && a != "&" && (s += (s && "-") + a);
    }
    s && (l && (s = "-" + s), t[0].push({
      n: s,
      v: o.filter(Ke),
      i: n
    }));
  }
}
function Ke(e, t, r) {
  return r.indexOf(e) == t;
}
let Ee = /* @__PURE__ */ new Map();
function oe(e) {
  let t = Ee.get(e);
  if (!t) {
    let r = [], o = [
      []
    ], n = 0, l = 0, s = null, a = 0, i = (f, p = 0) => {
      n != a && (r.push(e.slice(n, a + p)), f && ce(r, o)), n = a + 1;
    };
    for (; a < e.length; a++) {
      let f = e[a];
      if (l) e[a - 1] != "\\" && (l += +(f == "[") || -(f == "]"));
      else if (f == "[")
        l += 1;
      else if (s)
        e[a - 1] != "\\" && s.test(e.slice(a)) && (s = null, n = a + RegExp.lastMatch.length);
      else if (f == "/" && e[a - 1] != "\\" && (e[a + 1] == "*" || e[a + 1] == "/"))
        s = e[a + 1] == "*" ? /^\*\// : /^[\r\n]/;
      else if (f == "(")
        i(), r.push(f);
      else if (f == ":") e[a + 1] != ":" && i(!1, 1);
      else if (/[\s,)]/.test(f)) {
        i(!0);
        let p = r.lastIndexOf("(");
        if (f == ")") {
          let u = r[p - 1];
          if (/[~@]$/.test(u)) {
            let g = o.shift();
            r.length = p, ce([
              ...r,
              "#"
            ], o);
            let { v: y } = o[0].pop();
            for (let x of g)
              x.v.splice(+(x.v[0] == "dark") - +(y[0] == "dark"), y.length);
            ce([
              ...r,
              Qe(
                // named nested
                u.length > 1 ? u.slice(0, -1) + te(JSON.stringify([
                  u,
                  g
                ])) : u + "(" + Xe(g) + ")",
                T.a,
                g,
                /@$/.test(u)
              )
            ], o);
          }
          p = r.lastIndexOf("(", p - 1);
        }
        r.length = p + 1;
      } else /[~@]/.test(f) && e[a + 1] == "(" && // start nested block
      // ~(...) or button~(...)
      // @(...) or button@(...)
      o.unshift([]);
    }
    i(!0), Ee.set(e, t = o[0]);
  }
  return t;
}
function c(e, t, r) {
  return [
    e,
    me(t, r)
  ];
}
function me(e, t) {
  return typeof e == "function" ? e : typeof e == "string" && /^[\w-]+$/.test(e) ? (
    // a CSS property alias
    (r, o) => ({
      [e]: t ? t(r, o) : be(r, 1)
    })
  ) : (r) => (
    // CSSObject, shortcut or apply
    e || {
      [r[1]]: be(r, 2)
    }
  );
}
function be(e, t, r = e.slice(t).find(Boolean) || e.$$ || e.input) {
  return e.input[0] == "-" ? `calc(${r} * -1)` : r;
}
function d(e, t, r, o) {
  return [
    e,
    et(t, r, o)
  ];
}
function et(e, t, r) {
  let o = typeof t == "string" ? (n, l) => ({
    [t]: r ? r(n, l) : n._
  }) : t || (({ 1: n, _: l }, s, a) => ({
    [n || a]: l
  }));
  return (n, l) => {
    let s = qe(e || n[1]), a = l.theme(s, n.$$) ?? I(n.$$, s, l);
    if (a != null) return n._ = be(n, 0, a), o(n, l, s);
  };
}
function S(e, t = {}, r) {
  return [
    e,
    tt(t, r)
  ];
}
function tt(e = {}, t) {
  return (r, o) => {
    let { section: n = qe(r[0]).replace("-", "") + "Color" } = e, [l, s] = rt(r.$$);
    if (!l) return;
    let a = o.theme(n, l) || I(l, n, o);
    if (!a || typeof a == "object") return;
    let {
      // text- -> --tw-text-opacity
      // ring-offset(?:-|$) -> --tw-ring-offset-opacity
      // TODO move this default into preset-tailwind?
      opacityVariable: i = `--tw-${r[0].replace(/-$/, "")}-opacity`,
      opacitySection: f = n.replace("Color", "Opacity"),
      property: p = n,
      selector: u
    } = e, g = o.theme(f, s || "DEFAULT") || s && I(s, f, o), y = t || (({ _: v }) => {
      let $ = ee(p, v);
      return u ? {
        [u]: $
      } : $;
    });
    r._ = {
      value: U(a, {
        opacityVariable: i || void 0,
        opacityValue: g || void 0
      }),
      color: (v) => U(a, v),
      opacityVariable: i || void 0,
      opacityValue: g || void 0
    };
    let x = y(r, o);
    if (!r.dark) {
      let v = o.d(n, l, a);
      v && v !== a && (r._ = {
        value: U(v, {
          opacityVariable: i || void 0,
          opacityValue: g || "1"
        }),
        color: ($) => U(v, $),
        opacityVariable: i || void 0,
        opacityValue: g || void 0
      }, x = {
        "&": x,
        [o.v("dark")]: y(r, o)
      });
    }
    return x;
  };
}
function rt(e) {
  return (e.match(/^(\[[^\]]+]|[^/]+?)(?:\/(.+))?$/) || []).slice(1);
}
function ee(e, t) {
  let r = {};
  return typeof t == "string" ? r[e] = t : (t.opacityVariable && t.value.includes(t.opacityVariable) && (r[t.opacityVariable] = t.opacityValue || "1"), r[e] = t.value), r;
}
function I(e, t, r) {
  if (e[0] == "[" && e.slice(-1) == "]") {
    if (e = Q(Ce(e.slice(1, -1), r.theme)), !t) return e;
    if (
      // Respect type hints from the user on ambiguous arbitrary values - https://tailwindcss.com/docs/adding-custom-styles#resolving-ambiguities
      !// If this is a color section and the value is a hex color, color function or color name
      (/color|fill|stroke/i.test(t) && !(/^color:/.test(e) || /^(#|((hsl|rgb)a?|hwb|lab|lch|color)\(|[a-z]+$)/.test(e)) || // url(, [a-z]-gradient(, image(, cross-fade(, image-set(
      /image/i.test(t) && !(/^image:/.test(e) || /^[a-z-]+\(/.test(e)) || // font-*
      // - fontWeight (type: ['lookup', 'number', 'any'])
      // - fontFamily (type: ['lookup', 'generic-name', 'family-name'])
      /weight/i.test(t) && !(/^(number|any):/.test(e) || /^\d+$/.test(e)) || // bg-*
      // - backgroundPosition (type: ['lookup', ['position', { preferOnConflict: true }]])
      // - backgroundSize (type: ['lookup', 'length', 'percentage', 'size'])
      /position/i.test(t) && /^(length|size):/.test(e))
    )
      return e.replace(/^[a-z-]+:/, "");
  }
}
function qe(e) {
  return e.replace(/-./g, (t) => t[1].toUpperCase());
}
function Q(e) {
  return (
    // Keep raw strings if it starts with `url(`
    e.includes("url(") ? e.replace(/(.*?)(url\(.*?\))(.*?)/g, (t, r = "", o, n = "") => Q(r) + o + Q(n)) : e.replace(/(^|[^\\])_+/g, (t, r) => r + " ".repeat(t.length - r.length)).replace(/\\_/g, "_").replace(/(calc|min|max|clamp)\(.+\)/g, (t) => t.replace(/(-?\d*\.?\d(?!\b-.+[,)](?![^+\-/*])\D)(?:%|[a-z]+)?|\))([+\-/*])/g, "$1 $2 "))
  );
}
function Ge({ presets: e = [], ...t }) {
  let r = {
    darkMode: void 0,
    darkColor: void 0,
    preflight: t.preflight !== !1 && [],
    theme: {},
    variants: w(t.variants),
    rules: w(t.rules),
    ignorelist: w(t.ignorelist),
    hash: void 0,
    stringify: (o, n) => o + ":" + n,
    finalize: []
  };
  for (let o of w([
    ...e,
    {
      darkMode: t.darkMode,
      darkColor: t.darkColor,
      preflight: t.preflight !== !1 && w(t.preflight),
      theme: t.theme,
      hash: t.hash,
      stringify: t.stringify,
      finalize: t.finalize
    }
  ])) {
    let { preflight: n, darkMode: l = r.darkMode, darkColor: s = r.darkColor, theme: a, variants: i, rules: f, ignorelist: p, hash: u = r.hash, stringify: g = r.stringify, finalize: y } = typeof o == "function" ? o(r) : o;
    r = {
      // values defined by user or previous presets take precedence
      preflight: r.preflight !== !1 && n !== !1 && [
        ...r.preflight,
        ...w(n)
      ],
      darkMode: l,
      darkColor: s,
      theme: {
        ...r.theme,
        ...a,
        extend: {
          ...r.theme.extend,
          ...a == null ? void 0 : a.extend
        }
      },
      variants: [
        ...r.variants,
        ...w(i)
      ],
      rules: [
        ...r.rules,
        ...w(f)
      ],
      ignorelist: [
        ...r.ignorelist,
        ...w(p)
      ],
      hash: u,
      stringify: g,
      finalize: [
        ...r.finalize,
        ...w(y)
      ]
    };
  }
  return r;
}
function Oe(e, t, r, o, n, l) {
  for (let s of t) {
    let a = r.get(s);
    a || r.set(s, a = o(s));
    let i = a(e, n, l);
    if (i) return i;
  }
}
function ot(e) {
  var t;
  return he(e[0], typeof (t = e[1]) == "function" ? t : () => t);
}
function nt(e) {
  var t, r;
  return Array.isArray(e) ? he(e[0], me(e[1], e[2])) : he(e, me(t, r));
}
function he(e, t) {
  return Ye(e, (r, o, n, l) => {
    let s = o.exec(r);
    if (s) return (
      // MATCH.$_ = value
      s.$$ = r.slice(s[0].length), s.dark = l, t(s, n)
    );
  });
}
function Ye(e, t) {
  let r = w(e).map(it);
  return (o, n, l) => {
    for (let s of r) {
      let a = t(o, s, n, l);
      if (a) return a;
    }
  };
}
function it(e) {
  return typeof e == "string" ? RegExp("^" + e + (e.includes("$") || e.slice(-1) == "-" ? "" : "$")) : e;
}
function at(e, t) {
  let r = Ge(e), o = function({ theme: i, darkMode: f, darkColor: p = ke, variants: u, rules: g, hash: y, stringify: x, ignorelist: v, finalize: $ }) {
    let B = /* @__PURE__ */ new Map(), b = /* @__PURE__ */ new Map(), k = /* @__PURE__ */ new Map(), z = /* @__PURE__ */ new Map(), ne = Ye(v, (h, R) => R.test(h));
    u.push([
      "dark",
      Array.isArray(f) || f == "class" ? `${w(f)[1] || ".dark"} &` : typeof f == "string" && f != "media" ? f : (
        // a custom selector
        "@media (prefers-color-scheme:dark)"
      )
    ]);
    let q = typeof y == "function" ? (h) => y(h, te) : y ? te : Te;
    q !== Te && $.push((h) => {
      var R;
      return {
        ...h,
        n: h.n && q(h.n),
        d: (R = h.d) == null ? void 0 : R.replace(/--(tw(?:-[\w-]+)?)\b/g, (j, ie) => "--" + q(ie).replace("#", ""))
      };
    });
    let M = {
      theme: function({ extend: h = {}, ...R }) {
        let j = {}, ie = {
          get colors() {
            return Y("colors");
          },
          theme: Y,
          // Stub implementation as negated values are automatically infered and do _not_ need to be in the theme
          negative() {
            return {};
          },
          breakpoints(C) {
            let A = {};
            for (let F in C) typeof C[F] == "string" && (A["screen-" + F] = C[F]);
            return A;
          }
        };
        return Y;
        function Y(C, A, F, _) {
          if (C) {
            if ({ 1: C, 2: _ } = // eslint-disable-next-line no-sparse-arrays
            /^(\S+?)(?:\s*\/\s*([^/]+))?$/.exec(C) || [
              ,
              C
            ], /[.[]/.test(C)) {
              let D = [];
              C.replace(/\[([^\]]+)\]|([^.[]+)/g, (N, X, Je = X) => D.push(Je)), C = D.shift(), F = A, A = D.join("-");
            }
            let V = j[C] || // two-step deref to allow extend section to reference base section
            Object.assign(Object.assign(
              // Make sure to not get into recursive calls
              j[C] = {},
              Re(R, C)
            ), Re(h, C));
            if (A == null) return V;
            A || (A = "DEFAULT");
            let H = V[A] ?? A.split("-").reduce((D, N) => D == null ? void 0 : D[N], V) ?? F;
            return _ ? U(H, {
              opacityValue: Ce(_, Y)
            }) : H;
          }
          let J = {};
          for (let V of [
            ...Object.keys(R),
            ...Object.keys(h)
          ]) J[V] = Y(V);
          return J;
        }
        function Re(C, A) {
          let F = C[A];
          return typeof F == "function" && (F = F(ie)), F && /color|fill|stroke/i.test(A) ? function _(J, V = []) {
            let H = {};
            for (let D in J) {
              let N = J[D], X = [
                ...V,
                D
              ];
              H[X.join("-")] = N, D == "DEFAULT" && (X = V, H[V.join("-")] = N), typeof N == "object" && Object.assign(H, _(N, X));
            }
            return H;
          }(F) : F;
        }
      }(i),
      e: Le,
      h: q,
      s(h, R) {
        return x(h, R, M);
      },
      d(h, R, j) {
        return p(h, R, M, j);
      },
      v(h) {
        return B.has(h) || B.set(h, Oe(h, u, b, ot, M) || "&:" + h), B.get(h);
      },
      r(h, R) {
        let j = JSON.stringify([
          h,
          R
        ]);
        return k.has(j) || k.set(j, !ne(h, M) && Oe(h, g, z, nt, M, R)), k.get(j);
      },
      f(h) {
        return $.reduce((R, j) => j(R, M), h);
      }
    };
    return M;
  }(r), n = /* @__PURE__ */ new Map(), l = [], s = /* @__PURE__ */ new Set();
  t.resume((i) => n.set(i, i), (i, f) => {
    t.insert(i, l.length, f), l.push(f), s.add(i);
  });
  function a(i) {
    let f = o.f(i), p = ge(f);
    if (p && !s.has(p)) {
      s.add(p);
      let u = Be(l, i);
      t.insert(p, u, i), l.splice(u, 0, i);
    }
    return f.n;
  }
  return Object.defineProperties(function(f) {
    if (!n.size) for (let u of w(r.preflight))
      typeof u == "function" && (u = u(o)), u && (typeof u == "string" ? $e("", T.b, oe(u), o, T.b, [], !1, !0) : Ne(u, {}, o, T.b)).forEach(a);
    f = "" + f;
    let p = n.get(f);
    if (!p) {
      let u = /* @__PURE__ */ new Set();
      for (let g of re(oe(f), o)) u.add(g.c).add(a(g));
      p = [
        ...u
      ].filter(Boolean).join(" "), n.set(f, p).set(p, p);
    }
    return p;
  }, Object.getOwnPropertyDescriptors({
    get target() {
      return t.target;
    },
    theme: o.theme,
    config: r,
    snapshot() {
      let i = t.snapshot(), f = new Set(s), p = new Map(n), u = [
        ...l
      ];
      return () => {
        i(), s = f, n = p, l = u;
      };
    },
    clear() {
      t.clear(), s = /* @__PURE__ */ new Set(), n = /* @__PURE__ */ new Map(), l = [];
    },
    destroy() {
      this.clear(), t.destroy();
    }
  }));
}
function lt(e, t) {
  return e != t && "" + e.split(" ").sort() != "" + t.split(" ").sort();
}
function st(e) {
  let t = new MutationObserver(r);
  return {
    observe(n) {
      t.observe(n, {
        attributeFilter: [
          "class"
        ],
        subtree: !0,
        childList: !0
      }), o(n), r([
        {
          target: n,
          type: ""
        }
      ]);
    },
    disconnect() {
      t.disconnect();
    }
  };
  function r(n) {
    for (let { type: l, target: s } of n) if (l[0] == "a")
      o(s);
    else
      for (let a of s.querySelectorAll("[class]")) o(a);
    t.takeRecords();
  }
  function o(n) {
    var a;
    let l, s = (a = n.getAttribute) == null ? void 0 : a.call(n, "class");
    s && lt(s, l = e(s)) && // Not using `target.className = ...` as that is read-only for SVGElements
    n.setAttribute("class", l);
  }
}
function ct(e) {
  let t = document.querySelector(e || 'style[data-twind=""]');
  return (!t || t.tagName != "STYLE") && (t = document.createElement("style"), document.head.prepend(t)), t.dataset.twind = "claimed", t;
}
function de(e) {
  let t = e != null && e.cssRules ? e : (e && typeof e != "string" ? e : ct(e)).sheet;
  return {
    target: t,
    snapshot() {
      let r = Array.from(t.cssRules, (o) => o.cssText);
      return () => {
        this.clear(), r.forEach(this.insert);
      };
    },
    clear() {
      for (let r = t.cssRules.length; r--; ) t.deleteRule(r);
    },
    destroy() {
      var r;
      (r = t.ownerNode) == null || r.remove();
    },
    insert(r, o) {
      try {
        t.insertRule(r, o);
      } catch {
        t.insertRule(":root{}", o);
      }
    },
    resume: ke
  };
}
function dt(e, t = !0) {
  let r = function() {
    if (ft) try {
      let i = de(new CSSStyleSheet());
      return i.connect = (f) => {
        let p = fe(f);
        p.adoptedStyleSheets = [
          ...p.adoptedStyleSheets,
          i.target
        ];
      }, i.disconnect = ke, i;
    } catch {
    }
    let l = document.createElement("style");
    l.media = "not all", document.head.prepend(l);
    let s = [
      de(l)
    ], a = /* @__PURE__ */ new WeakMap();
    return {
      get target() {
        return s[0].target;
      },
      snapshot() {
        let i = s.map((f) => f.snapshot());
        return () => i.forEach((f) => f());
      },
      clear() {
        s.forEach((i) => i.clear());
      },
      destroy() {
        s.forEach((i) => i.destroy());
      },
      insert(i, f, p) {
        s[0].insert(i, f, p);
        let u = this.target.cssRules[f];
        s.forEach((g, y) => y && g.target.insertRule(u.cssText, f));
      },
      resume(i, f) {
        return s[0].resume(i, f);
      },
      connect(i) {
        let f = document.createElement("style");
        fe(i).appendChild(f);
        let p = de(f), { cssRules: u } = this.target;
        for (let g = 0; g < u.length; g++) p.target.insertRule(u[g].cssText, g);
        s.push(p), a.set(i, p);
      },
      disconnect(i) {
        let f = s.indexOf(a.get(i));
        f >= 0 && s.splice(f, 1);
      }
    };
  }(), o = at({
    ...e,
    // in production use short hashed class names
    hash: e.hash ?? t
  }, r), n = st(o);
  return function(s) {
    return class extends s {
      connectedCallback() {
        var i;
        (i = super.connectedCallback) == null || i.call(this), r.connect(this), n.observe(fe(this));
      }
      disconnectedCallback() {
        var i;
        r.disconnect(this), (i = super.disconnectedCallback) == null || i.call(this);
      }
      constructor(...i) {
        super(...i), this.tw = o;
      }
    };
  };
}
let ft = typeof ShadowRoot < "u" && (typeof ShadyCSS > "u" || ShadyCSS.nativeShadow) && "adoptedStyleSheets" in Document.prototype && "replace" in CSSStyleSheet.prototype;
function fe(e) {
  return e.shadowRoot || e.attachShadow({
    mode: "open"
  });
}
var pt = /* @__PURE__ */ new Map([["align-self", "-ms-grid-row-align"], ["color-adjust", "-webkit-print-color-adjust"], ["column-gap", "grid-column-gap"], ["forced-color-adjust", "-ms-high-contrast-adjust"], ["gap", "grid-gap"], ["grid-template-columns", "-ms-grid-columns"], ["grid-template-rows", "-ms-grid-rows"], ["justify-self", "-ms-grid-column-align"], ["margin-inline-end", "-webkit-margin-end"], ["margin-inline-start", "-webkit-margin-start"], ["mask-border", "-webkit-mask-box-image"], ["mask-border-outset", "-webkit-mask-box-image-outset"], ["mask-border-slice", "-webkit-mask-box-image-slice"], ["mask-border-source", "-webkit-mask-box-image-source"], ["mask-border-repeat", "-webkit-mask-box-image-repeat"], ["mask-border-width", "-webkit-mask-box-image-width"], ["overflow-wrap", "word-wrap"], ["padding-inline-end", "-webkit-padding-end"], ["padding-inline-start", "-webkit-padding-start"], ["print-color-adjust", "color-adjust"], ["row-gap", "grid-row-gap"], ["scroll-margin-bottom", "scroll-snap-margin-bottom"], ["scroll-margin-left", "scroll-snap-margin-left"], ["scroll-margin-right", "scroll-snap-margin-right"], ["scroll-margin-top", "scroll-snap-margin-top"], ["scroll-margin", "scroll-snap-margin"], ["text-combine-upright", "-ms-text-combine-horizontal"]]);
function ut(e) {
  return pt.get(e);
}
function gt(e) {
  var t = /^(?:(text-(?:decoration$|e|or|si)|back(?:ground-cl|d|f)|box-d|mask(?:$|-[ispro]|-cl)|pr|hyphena|flex-d)|(tab-|column(?!-s)|text-align-l)|(ap)|u|hy)/i.exec(e);
  return t ? t[1] ? 1 : t[2] ? 2 : t[3] ? 3 : 5 : 0;
}
function mt(e, t) {
  var r = /^(?:(pos)|(cli)|(background-i)|(flex(?:$|-b)|(?:max-|min-)?(?:block-s|inl|he|widt))|dis)/i.exec(e);
  return r ? r[1] ? /^sti/i.test(t) ? 1 : 0 : r[2] ? /^pat/i.test(t) ? 1 : 0 : r[3] ? /^image-/i.test(t) ? 1 : 0 : r[4] ? t[3] === "-" ? 2 : 0 : /^(?:inline-)?grid$/i.test(t) ? 4 : 0 : 0;
}
let bt = [
  [
    "-webkit-",
    1
  ],
  // 0b001
  [
    "-moz-",
    2
  ],
  // 0b010
  [
    "-ms-",
    4
  ]
];
function ht() {
  return ({ stringify: e }) => ({
    stringify(t, r, o) {
      let n = "", l = ut(t);
      l && (n += e(l, r, o) + ";");
      let s = gt(t), a = mt(t, r);
      for (let i of bt)
        s & i[1] && (n += e(i[0] + t, r, o) + ";"), a & i[1] && (n += e(t, i[0] + r, o) + ";");
      return n + e(t, r, o);
    }
  });
}
let we = {
  screens: {
    sm: "640px",
    md: "768px",
    lg: "1024px",
    xl: "1280px",
    "2xl": "1536px"
  },
  columns: {
    auto: "auto",
    // Handled by plugin,
    // 1: '1',
    // 2: '2',
    // 3: '3',
    // 4: '4',
    // 5: '5',
    // 6: '6',
    // 7: '7',
    // 8: '8',
    // 9: '9',
    // 10: '10',
    // 11: '11',
    // 12: '12',
    "3xs": "16rem",
    "2xs": "18rem",
    xs: "20rem",
    sm: "24rem",
    md: "28rem",
    lg: "32rem",
    xl: "36rem",
    "2xl": "42rem",
    "3xl": "48rem",
    "4xl": "56rem",
    "5xl": "64rem",
    "6xl": "72rem",
    "7xl": "80rem"
  },
  spacing: {
    px: "1px",
    0: "0px",
    .../* @__PURE__ */ E(4, "rem", 4, 0.5, 0.5),
    // 0.5: '0.125rem',
    // 1: '0.25rem',
    // 1.5: '0.375rem',
    // 2: '0.5rem',
    // 2.5: '0.625rem',
    // 3: '0.75rem',
    // 3.5: '0.875rem',
    // 4: '1rem',
    .../* @__PURE__ */ E(12, "rem", 4, 5),
    // 5: '1.25rem',
    // 6: '1.5rem',
    // 7: '1.75rem',
    // 8: '2rem',
    // 9: '2.25rem',
    // 10: '2.5rem',
    // 11: '2.75rem',
    // 12: '3rem',
    14: "3.5rem",
    .../* @__PURE__ */ E(64, "rem", 4, 16, 4),
    // 16: '4rem',
    // 20: '5rem',
    // 24: '6rem',
    // 28: '7rem',
    // 32: '8rem',
    // 36: '9rem',
    // 40: '10rem',
    // 44: '11rem',
    // 48: '12rem',
    // 52: '13rem',
    // 56: '14rem',
    // 60: '15rem',
    // 64: '16rem',
    72: "18rem",
    80: "20rem",
    96: "24rem"
  },
  durations: {
    75: "75ms",
    100: "100ms",
    150: "150ms",
    200: "200ms",
    300: "300ms",
    500: "500ms",
    700: "700ms",
    1e3: "1000ms"
  },
  animation: {
    none: "none",
    spin: "spin 1s linear infinite",
    ping: "ping 1s cubic-bezier(0,0,0.2,1) infinite",
    pulse: "pulse 2s cubic-bezier(0.4,0,0.6,1) infinite",
    bounce: "bounce 1s infinite"
  },
  aspectRatio: {
    auto: "auto",
    square: "1/1",
    video: "16/9"
  },
  backdropBlur: /* @__PURE__ */ m("blur"),
  backdropBrightness: /* @__PURE__ */ m("brightness"),
  backdropContrast: /* @__PURE__ */ m("contrast"),
  backdropGrayscale: /* @__PURE__ */ m("grayscale"),
  backdropHueRotate: /* @__PURE__ */ m("hueRotate"),
  backdropInvert: /* @__PURE__ */ m("invert"),
  backdropOpacity: /* @__PURE__ */ m("opacity"),
  backdropSaturate: /* @__PURE__ */ m("saturate"),
  backdropSepia: /* @__PURE__ */ m("sepia"),
  backgroundColor: /* @__PURE__ */ m("colors"),
  backgroundImage: {
    none: "none"
  },
  // These are built-in
  // 'gradient-to-t': 'linear-gradient(to top, var(--tw-gradient-stops))',
  // 'gradient-to-tr': 'linear-gradient(to top right, var(--tw-gradient-stops))',
  // 'gradient-to-r': 'linear-gradient(to right, var(--tw-gradient-stops))',
  // 'gradient-to-br': 'linear-gradient(to bottom right, var(--tw-gradient-stops))',
  // 'gradient-to-b': 'linear-gradient(to bottom, var(--tw-gradient-stops))',
  // 'gradient-to-bl': 'linear-gradient(to bottom left, var(--tw-gradient-stops))',
  // 'gradient-to-l': 'linear-gradient(to left, var(--tw-gradient-stops))',
  // 'gradient-to-tl': 'linear-gradient(to top left, var(--tw-gradient-stops))',
  backgroundOpacity: /* @__PURE__ */ m("opacity"),
  // backgroundPosition: {
  //   // The following are already handled by the plugin:
  //   // center, right, left, bottom, top
  //   // 'bottom-10px-right-20px' -> bottom 10px right 20px
  // },
  backgroundSize: {
    auto: "auto",
    cover: "cover",
    contain: "contain"
  },
  blur: {
    none: "none",
    0: "0",
    sm: "4px",
    DEFAULT: "8px",
    md: "12px",
    lg: "16px",
    xl: "24px",
    "2xl": "40px",
    "3xl": "64px"
  },
  brightness: {
    .../* @__PURE__ */ E(200, "", 100, 0, 50),
    // 0: '0',
    // 50: '.5',
    // 150: '1.5',
    // 200: '2',
    .../* @__PURE__ */ E(110, "", 100, 90, 5),
    // 90: '.9',
    // 95: '.95',
    // 100: '1',
    // 105: '1.05',
    // 110: '1.1',
    75: "0.75",
    125: "1.25"
  },
  borderColor: ({ theme: e }) => ({
    DEFAULT: e("colors.gray.200", "currentColor"),
    ...e("colors")
  }),
  borderOpacity: /* @__PURE__ */ m("opacity"),
  borderRadius: {
    none: "0px",
    sm: "0.125rem",
    DEFAULT: "0.25rem",
    md: "0.375rem",
    lg: "0.5rem",
    xl: "0.75rem",
    "2xl": "1rem",
    "3xl": "1.5rem",
    "1/2": "50%",
    full: "9999px"
  },
  borderSpacing: /* @__PURE__ */ m("spacing"),
  borderWidth: {
    DEFAULT: "1px",
    .../* @__PURE__ */ O(8, "px")
  },
  // 0: '0px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  boxShadow: {
    sm: "0 1px 2px 0 rgba(0,0,0,0.05)",
    DEFAULT: "0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px -1px rgba(0,0,0,0.1)",
    md: "0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -2px rgba(0,0,0,0.1)",
    lg: "0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1)",
    xl: "0 20px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1)",
    "2xl": "0 25px 50px -12px rgba(0,0,0,0.25)",
    inner: "inset 0 2px 4px 0 rgba(0,0,0,0.05)",
    none: "0 0 #0000"
  },
  boxShadowColor: m("colors"),
  // container: {},
  // cursor: {
  //   // Default values are handled by plugin
  // },
  caretColor: /* @__PURE__ */ m("colors"),
  accentColor: ({ theme: e }) => ({
    auto: "auto",
    ...e("colors")
  }),
  contrast: {
    .../* @__PURE__ */ E(200, "", 100, 0, 50),
    // 0: '0',
    // 50: '.5',
    // 150: '1.5',
    // 200: '2',
    75: "0.75",
    125: "1.25"
  },
  content: {
    none: "none"
  },
  divideColor: /* @__PURE__ */ m("borderColor"),
  divideOpacity: /* @__PURE__ */ m("borderOpacity"),
  divideWidth: /* @__PURE__ */ m("borderWidth"),
  dropShadow: {
    sm: "0 1px 1px rgba(0,0,0,0.05)",
    DEFAULT: [
      "0 1px 2px rgba(0,0,0,0.1)",
      "0 1px 1px rgba(0,0,0,0.06)"
    ],
    md: [
      "0 4px 3px rgba(0,0,0,0.07)",
      "0 2px 2px rgba(0,0,0,0.06)"
    ],
    lg: [
      "0 10px 8px rgba(0,0,0,0.04)",
      "0 4px 3px rgba(0,0,0,0.1)"
    ],
    xl: [
      "0 20px 13px rgba(0,0,0,0.03)",
      "0 8px 5px rgba(0,0,0,0.08)"
    ],
    "2xl": "0 25px 25px rgba(0,0,0,0.15)",
    none: "0 0 #0000"
  },
  fill: ({ theme: e }) => ({
    ...e("colors"),
    none: "none"
  }),
  grayscale: {
    DEFAULT: "100%",
    0: "0"
  },
  hueRotate: {
    0: "0deg",
    15: "15deg",
    30: "30deg",
    60: "60deg",
    90: "90deg",
    180: "180deg"
  },
  invert: {
    DEFAULT: "100%",
    0: "0"
  },
  flex: {
    1: "1 1 0%",
    auto: "1 1 auto",
    initial: "0 1 auto",
    none: "none"
  },
  flexBasis: ({ theme: e }) => ({
    ...e("spacing"),
    ...Z(2, 6),
    // '1/2': '50%',
    // '1/3': '33.333333%',
    // '2/3': '66.666667%',
    // '1/4': '25%',
    // '2/4': '50%',
    // '3/4': '75%',
    // '1/5': '20%',
    // '2/5': '40%',
    // '3/5': '60%',
    // '4/5': '80%',
    // '1/6': '16.666667%',
    // '2/6': '33.333333%',
    // '3/6': '50%',
    // '4/6': '66.666667%',
    // '5/6': '83.333333%',
    ...Z(12, 12),
    // '1/12': '8.333333%',
    // '2/12': '16.666667%',
    // '3/12': '25%',
    // '4/12': '33.333333%',
    // '5/12': '41.666667%',
    // '6/12': '50%',
    // '7/12': '58.333333%',
    // '8/12': '66.666667%',
    // '9/12': '75%',
    // '10/12': '83.333333%',
    // '11/12': '91.666667%',
    auto: "auto",
    full: "100%"
  }),
  flexGrow: {
    DEFAULT: 1,
    0: 0
  },
  flexShrink: {
    DEFAULT: 1,
    0: 0
  },
  fontFamily: {
    sans: 'ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,"Noto Sans",sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji"'.split(","),
    serif: 'ui-serif,Georgia,Cambria,"Times New Roman",Times,serif'.split(","),
    mono: 'ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace'.split(",")
  },
  fontSize: {
    xs: [
      "0.75rem",
      "1rem"
    ],
    sm: [
      "0.875rem",
      "1.25rem"
    ],
    base: [
      "1rem",
      "1.5rem"
    ],
    lg: [
      "1.125rem",
      "1.75rem"
    ],
    xl: [
      "1.25rem",
      "1.75rem"
    ],
    "2xl": [
      "1.5rem",
      "2rem"
    ],
    "3xl": [
      "1.875rem",
      "2.25rem"
    ],
    "4xl": [
      "2.25rem",
      "2.5rem"
    ],
    "5xl": [
      "3rem",
      "1"
    ],
    "6xl": [
      "3.75rem",
      "1"
    ],
    "7xl": [
      "4.5rem",
      "1"
    ],
    "8xl": [
      "6rem",
      "1"
    ],
    "9xl": [
      "8rem",
      "1"
    ]
  },
  fontWeight: {
    thin: "100",
    extralight: "200",
    light: "300",
    normal: "400",
    medium: "500",
    semibold: "600",
    bold: "700",
    extrabold: "800",
    black: "900"
  },
  gap: /* @__PURE__ */ m("spacing"),
  gradientColorStops: /* @__PURE__ */ m("colors"),
  gridAutoColumns: {
    auto: "auto",
    min: "min-content",
    max: "max-content",
    fr: "minmax(0,1fr)"
  },
  gridAutoRows: {
    auto: "auto",
    min: "min-content",
    max: "max-content",
    fr: "minmax(0,1fr)"
  },
  gridColumn: {
    // span-X is handled by the plugin: span-1 -> span 1 / span 1
    auto: "auto",
    "span-full": "1 / -1"
  },
  // gridColumnEnd: {
  //   // Defaults handled by plugin
  // },
  // gridColumnStart: {
  //   // Defaults handled by plugin
  // },
  gridRow: {
    // span-X is handled by the plugin: span-1 -> span 1 / span 1
    auto: "auto",
    "span-full": "1 / -1"
  },
  // gridRowStart: {
  //   // Defaults handled by plugin
  // },
  // gridRowEnd: {
  //   // Defaults handled by plugin
  // },
  gridTemplateColumns: {
    // numbers are handled by the plugin: 1 -> repeat(1, minmax(0, 1fr))
    none: "none"
  },
  gridTemplateRows: {
    // numbers are handled by the plugin: 1 -> repeat(1, minmax(0, 1fr))
    none: "none"
  },
  height: ({ theme: e }) => ({
    ...e("spacing"),
    ...Z(2, 6),
    // '1/2': '50%',
    // '1/3': '33.333333%',
    // '2/3': '66.666667%',
    // '1/4': '25%',
    // '2/4': '50%',
    // '3/4': '75%',
    // '1/5': '20%',
    // '2/5': '40%',
    // '3/5': '60%',
    // '4/5': '80%',
    // '1/6': '16.666667%',
    // '2/6': '33.333333%',
    // '3/6': '50%',
    // '4/6': '66.666667%',
    // '5/6': '83.333333%',
    min: "min-content",
    max: "max-content",
    fit: "fit-content",
    auto: "auto",
    full: "100%",
    screen: "100vh"
  }),
  inset: ({ theme: e }) => ({
    ...e("spacing"),
    ...Z(2, 4),
    // '1/2': '50%',
    // '1/3': '33.333333%',
    // '2/3': '66.666667%',
    // '1/4': '25%',
    // '2/4': '50%',
    // '3/4': '75%',
    auto: "auto",
    full: "100%"
  }),
  keyframes: {
    spin: {
      from: {
        transform: "rotate(0deg)"
      },
      to: {
        transform: "rotate(360deg)"
      }
    },
    ping: {
      "0%": {
        transform: "scale(1)",
        opacity: "1"
      },
      "75%,100%": {
        transform: "scale(2)",
        opacity: "0"
      }
    },
    pulse: {
      "0%,100%": {
        opacity: "1"
      },
      "50%": {
        opacity: ".5"
      }
    },
    bounce: {
      "0%, 100%": {
        transform: "translateY(-25%)",
        animationTimingFunction: "cubic-bezier(0.8,0,1,1)"
      },
      "50%": {
        transform: "none",
        animationTimingFunction: "cubic-bezier(0,0,0.2,1)"
      }
    }
  },
  letterSpacing: {
    tighter: "-0.05em",
    tight: "-0.025em",
    normal: "0em",
    wide: "0.025em",
    wider: "0.05em",
    widest: "0.1em"
  },
  lineHeight: {
    .../* @__PURE__ */ E(10, "rem", 4, 3),
    // 3: '.75rem',
    // 4: '1rem',
    // 5: '1.25rem',
    // 6: '1.5rem',
    // 7: '1.75rem',
    // 8: '2rem',
    // 9: '2.25rem',
    // 10: '2.5rem',
    none: "1",
    tight: "1.25",
    snug: "1.375",
    normal: "1.5",
    relaxed: "1.625",
    loose: "2"
  },
  // listStyleType: {
  //   // Defaults handled by plugin
  // },
  margin: ({ theme: e }) => ({
    auto: "auto",
    ...e("spacing")
  }),
  maxHeight: ({ theme: e }) => ({
    full: "100%",
    min: "min-content",
    max: "max-content",
    fit: "fit-content",
    screen: "100vh",
    ...e("spacing")
  }),
  maxWidth: ({ theme: e, breakpoints: t }) => ({
    ...t(e("screens")),
    none: "none",
    0: "0rem",
    xs: "20rem",
    sm: "24rem",
    md: "28rem",
    lg: "32rem",
    xl: "36rem",
    "2xl": "42rem",
    "3xl": "48rem",
    "4xl": "56rem",
    "5xl": "64rem",
    "6xl": "72rem",
    "7xl": "80rem",
    full: "100%",
    min: "min-content",
    max: "max-content",
    fit: "fit-content",
    prose: "65ch"
  }),
  minHeight: {
    0: "0px",
    full: "100%",
    min: "min-content",
    max: "max-content",
    fit: "fit-content",
    screen: "100vh"
  },
  minWidth: {
    0: "0px",
    full: "100%",
    min: "min-content",
    max: "max-content",
    fit: "fit-content"
  },
  // objectPosition: {
  //   // The plugins joins all arguments by default
  // },
  opacity: {
    .../* @__PURE__ */ E(100, "", 100, 0, 10),
    // 0: '0',
    // 10: '0.1',
    // 20: '0.2',
    // 30: '0.3',
    // 40: '0.4',
    // 60: '0.6',
    // 70: '0.7',
    // 80: '0.8',
    // 90: '0.9',
    // 100: '1',
    5: "0.05",
    25: "0.25",
    75: "0.75",
    95: "0.95"
  },
  order: {
    // Handled by plugin
    // 1: '1',
    // 2: '2',
    // 3: '3',
    // 4: '4',
    // 5: '5',
    // 6: '6',
    // 7: '7',
    // 8: '8',
    // 9: '9',
    // 10: '10',
    // 11: '11',
    // 12: '12',
    first: "-9999",
    last: "9999",
    none: "0"
  },
  padding: /* @__PURE__ */ m("spacing"),
  placeholderColor: /* @__PURE__ */ m("colors"),
  placeholderOpacity: /* @__PURE__ */ m("opacity"),
  outlineColor: /* @__PURE__ */ m("colors"),
  outlineOffset: /* @__PURE__ */ O(8, "px"),
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',,
  outlineWidth: /* @__PURE__ */ O(8, "px"),
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',,
  ringColor: ({ theme: e }) => ({
    ...e("colors"),
    DEFAULT: "#3b82f6"
  }),
  ringOffsetColor: /* @__PURE__ */ m("colors"),
  ringOffsetWidth: /* @__PURE__ */ O(8, "px"),
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',,
  ringOpacity: ({ theme: e }) => ({
    ...e("opacity"),
    DEFAULT: "0.5"
  }),
  ringWidth: {
    DEFAULT: "3px",
    .../* @__PURE__ */ O(8, "px")
  },
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  rotate: {
    .../* @__PURE__ */ O(2, "deg"),
    // 0: '0deg',
    // 1: '1deg',
    // 2: '2deg',
    .../* @__PURE__ */ O(12, "deg", 3),
    // 3: '3deg',
    // 6: '6deg',
    // 12: '12deg',
    .../* @__PURE__ */ O(180, "deg", 45)
  },
  // 45: '45deg',
  // 90: '90deg',
  // 180: '180deg',
  saturate: /* @__PURE__ */ E(200, "", 100, 0, 50),
  // 0: '0',
  // 50: '.5',
  // 100: '1',
  // 150: '1.5',
  // 200: '2',
  scale: {
    .../* @__PURE__ */ E(150, "", 100, 0, 50),
    // 0: '0',
    // 50: '.5',
    // 150: '1.5',
    .../* @__PURE__ */ E(110, "", 100, 90, 5),
    // 90: '.9',
    // 95: '.95',
    // 100: '1',
    // 105: '1.05',
    // 110: '1.1',
    75: "0.75",
    125: "1.25"
  },
  scrollMargin: /* @__PURE__ */ m("spacing"),
  scrollPadding: /* @__PURE__ */ m("spacing"),
  sepia: {
    0: "0",
    DEFAULT: "100%"
  },
  skew: {
    .../* @__PURE__ */ O(2, "deg"),
    // 0: '0deg',
    // 1: '1deg',
    // 2: '2deg',
    .../* @__PURE__ */ O(12, "deg", 3)
  },
  // 3: '3deg',
  // 6: '6deg',
  // 12: '12deg',
  space: /* @__PURE__ */ m("spacing"),
  stroke: ({ theme: e }) => ({
    ...e("colors"),
    none: "none"
  }),
  strokeWidth: /* @__PURE__ */ E(2),
  // 0: '0',
  // 1: '1',
  // 2: '2',,
  textColor: /* @__PURE__ */ m("colors"),
  textDecorationColor: /* @__PURE__ */ m("colors"),
  textDecorationThickness: {
    "from-font": "from-font",
    auto: "auto",
    .../* @__PURE__ */ O(8, "px")
  },
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  textUnderlineOffset: {
    auto: "auto",
    .../* @__PURE__ */ O(8, "px")
  },
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  textIndent: /* @__PURE__ */ m("spacing"),
  textOpacity: /* @__PURE__ */ m("opacity"),
  // transformOrigin: {
  //   // The following are already handled by the plugin:
  //   // center, right, left, bottom, top
  //   // 'bottom-10px-right-20px' -> bottom 10px right 20px
  // },
  transitionDuration: ({ theme: e }) => ({
    ...e("durations"),
    DEFAULT: "150ms"
  }),
  transitionDelay: /* @__PURE__ */ m("durations"),
  transitionProperty: {
    none: "none",
    all: "all",
    DEFAULT: "color,background-color,border-color,text-decoration-color,fill,stroke,opacity,box-shadow,transform,filter,backdrop-filter",
    colors: "color,background-color,border-color,text-decoration-color,fill,stroke",
    opacity: "opacity",
    shadow: "box-shadow",
    transform: "transform"
  },
  transitionTimingFunction: {
    DEFAULT: "cubic-bezier(0.4,0,0.2,1)",
    linear: "linear",
    in: "cubic-bezier(0.4,0,1,1)",
    out: "cubic-bezier(0,0,0.2,1)",
    "in-out": "cubic-bezier(0.4,0,0.2,1)"
  },
  translate: ({ theme: e }) => ({
    ...e("spacing"),
    ...Z(2, 4),
    // '1/2': '50%',
    // '1/3': '33.333333%',
    // '2/3': '66.666667%',
    // '1/4': '25%',
    // '2/4': '50%',
    // '3/4': '75%',
    full: "100%"
  }),
  width: ({ theme: e }) => ({
    min: "min-content",
    max: "max-content",
    fit: "fit-content",
    screen: "100vw",
    ...e("flexBasis")
  }),
  willChange: {
    scroll: "scroll-position"
  },
  // other options handled by rules
  // auto: 'auto',
  // contents: 'contents',
  // transform: 'transform',
  zIndex: {
    .../* @__PURE__ */ E(50, "", 1, 0, 10),
    // 0: '0',
    // 10: '10',
    // 20: '20',
    // 30: '30',
    // 40: '40',
    // 50: '50',
    auto: "auto"
  }
};
function Z(e, t) {
  let r = {};
  do
    for (var o = 1; o < e; o++) r[`${o}/${e}`] = Number((o / e * 100).toFixed(6)) + "%";
  while (++e <= t);
  return r;
}
function O(e, t, r = 0) {
  let o = {};
  for (; r <= e; r = 2 * r || 1) o[r] = r + t;
  return o;
}
function E(e, t = "", r = 1, o = 0, n = 1, l = {}) {
  for (; o <= e; o += n) l[o] = o / r + t;
  return l;
}
function m(e) {
  return ({ theme: t }) => t(e);
}
let wt = {
  /*
  1. Prevent padding and border from affecting element width. (https://github.com/mozdevs/cssremedy/issues/4)
  2. Allow adding a border to an element by just adding a border-width. (https://github.com/tailwindcss/tailwindcss/pull/116)
  */
  "*,::before,::after": {
    boxSizing: "border-box",
    /* 1 */
    borderWidth: "0",
    /* 2 */
    borderStyle: "solid",
    /* 2 */
    borderColor: "theme(borderColor.DEFAULT, currentColor)"
  },
  /* 2 */
  "::before,::after": {
    "--tw-content": "''"
  },
  /*
  1. Use a consistent sensible line-height in all browsers.
  2. Prevent adjustments of font size after orientation changes in iOS.
  3. Use a more readable tab size.
  4. Use the user's configured `sans` font-family by default.
  5. Use the user's configured `sans` font-feature-settings by default.
  */
  html: {
    lineHeight: 1.5,
    /* 1 */
    WebkitTextSizeAdjust: "100%",
    /* 2 */
    MozTabSize: "4",
    /* 3 */
    tabSize: 4,
    /* 3 */
    fontFamily: `theme(fontFamily.sans, ${we.fontFamily.sans})`,
    /* 4 */
    fontFeatureSettings: "theme(fontFamily.sans[1].fontFeatureSettings, normal)"
  },
  /* 5 */
  /*
  1. Remove the margin in all browsers.
  2. Inherit line-height from `html` so users can set them as a class directly on the `html` element.
  */
  body: {
    margin: "0",
    /* 1 */
    lineHeight: "inherit"
  },
  /* 2 */
  /*
  1. Add the correct height in Firefox.
  2. Correct the inheritance of border color in Firefox. (https://bugzilla.mozilla.org/show_bug.cgi?id=190655)
  3. Ensure horizontal rules are visible by default.
  */
  hr: {
    height: "0",
    /* 1 */
    color: "inherit",
    /* 2 */
    borderTopWidth: "1px"
  },
  /* 3 */
  /*
  Add the correct text decoration in Chrome, Edge, and Safari.
  */
  "abbr:where([title])": {
    textDecoration: "underline dotted"
  },
  /*
  Remove the default font size and weight for headings.
  */
  "h1,h2,h3,h4,h5,h6": {
    fontSize: "inherit",
    fontWeight: "inherit"
  },
  /*
  Reset links to optimize for opt-in styling instead of opt-out.
  */
  a: {
    color: "inherit",
    textDecoration: "inherit"
  },
  /*
  Add the correct font weight in Edge and Safari.
  */
  "b,strong": {
    fontWeight: "bolder"
  },
  /*
  1. Use the user's configured `mono` font family by default.
  2. Use the user's configured `mono` font-feature-settings by default.
  3. Correct the odd `em` font sizing in all browsers.
  */
  "code,kbd,samp,pre": {
    fontFamily: `theme(fontFamily.mono, ${we.fontFamily.mono})`,
    fontFeatureSettings: "theme(fontFamily.mono[1].fontFeatureSettings, normal)",
    fontSize: "1em"
  },
  /*
  Add the correct font size in all browsers.
  */
  small: {
    fontSize: "80%"
  },
  /*
  Prevent `sub` and `sup` elements from affecting the line height in all browsers.
  */
  "sub,sup": {
    fontSize: "75%",
    lineHeight: 0,
    position: "relative",
    verticalAlign: "baseline"
  },
  sub: {
    bottom: "-0.25em"
  },
  sup: {
    top: "-0.5em"
  },
  /*
  1. Remove text indentation from table contents in Chrome and Safari. (https://bugs.chromium.org/p/chromium/issues/detail?id=999088, https://bugs.webkit.org/show_bug.cgi?id=201297)
  2. Correct table border color inheritance in all Chrome and Safari. (https://bugs.chromium.org/p/chromium/issues/detail?id=935729, https://bugs.webkit.org/show_bug.cgi?id=195016)
  3. Remove gaps between table borders by default.
  */
  table: {
    textIndent: "0",
    /* 1 */
    borderColor: "inherit",
    /* 2 */
    borderCollapse: "collapse"
  },
  /* 3 */
  /*
  1. Change the font styles in all browsers.
  2. Remove the margin in Firefox and Safari.
  3. Remove default padding in all browsers.
  */
  "button,input,optgroup,select,textarea": {
    fontFamily: "inherit",
    /* 1 */
    fontSize: "100%",
    /* 1 */
    lineHeight: "inherit",
    /* 1 */
    color: "inherit",
    /* 1 */
    margin: "0",
    /* 2 */
    padding: "0"
  },
  /* 3 */
  /*
  Remove the inheritance of text transform in Edge and Firefox.
  */
  "button,select": {
    textTransform: "none"
  },
  /*
  1. Correct the inability to style clickable types in iOS and Safari.
  2. Remove default button styles.
  */
  "button,[type='button'],[type='reset'],[type='submit']": {
    WebkitAppearance: "button",
    /* 1 */
    backgroundColor: "transparent",
    /* 2 */
    backgroundImage: "none"
  },
  /* 4 */
  /*
  Use the modern Firefox focus style for all focusable elements.
  */
  ":-moz-focusring": {
    outline: "auto"
  },
  /*
  Remove the additional `:invalid` styles in Firefox. (https://github.com/mozilla/gecko-dev/blob/2f9eacd9d3d995c937b4251a5557d95d494c9be1/layout/style/res/forms.css#L728-L737)
  */
  ":-moz-ui-invalid": {
    boxShadow: "none"
  },
  /*
  Add the correct vertical alignment in Chrome and Firefox.
  */
  progress: {
    verticalAlign: "baseline"
  },
  /*
  Correct the cursor style of increment and decrement buttons in Safari.
  */
  "::-webkit-inner-spin-button,::-webkit-outer-spin-button": {
    height: "auto"
  },
  /*
  1. Correct the odd appearance in Chrome and Safari.
  2. Correct the outline style in Safari.
  */
  "[type='search']": {
    WebkitAppearance: "textfield",
    /* 1 */
    outlineOffset: "-2px"
  },
  /* 2 */
  /*
  Remove the inner padding in Chrome and Safari on macOS.
  */
  "::-webkit-search-decoration": {
    WebkitAppearance: "none"
  },
  /*
  1. Correct the inability to style clickable types in iOS and Safari.
  2. Change font properties to `inherit` in Safari.
  */
  "::-webkit-file-upload-button": {
    WebkitAppearance: "button",
    /* 1 */
    font: "inherit"
  },
  /* 2 */
  /*
  Add the correct display in Chrome and Safari.
  */
  summary: {
    display: "list-item"
  },
  /*
  Removes the default spacing and border for appropriate elements.
  */
  "blockquote,dl,dd,h1,h2,h3,h4,h5,h6,hr,figure,p,pre": {
    margin: "0"
  },
  fieldset: {
    margin: "0",
    padding: "0"
  },
  legend: {
    padding: "0"
  },
  "ol,ul,menu": {
    listStyle: "none",
    margin: "0",
    padding: "0"
  },
  /*
  Prevent resizing textareas horizontally by default.
  */
  textarea: {
    resize: "vertical"
  },
  /*
  1. Reset the default placeholder opacity in Firefox. (https://github.com/tailwindlabs/tailwindcss/issues/3300)
  2. Set the default placeholder color to the user's configured gray 400 color.
  */
  "input::placeholder,textarea::placeholder": {
    opacity: 1,
    /* 1 */
    color: "theme(colors.gray.400, #9ca3af)"
  },
  /* 2 */
  /*
  Set the default cursor for buttons.
  */
  'button,[role="button"]': {
    cursor: "pointer"
  },
  /*
  Make sure disabled buttons don't get the pointer cursor.
  */
  ":disabled": {
    cursor: "default"
  },
  /*
  1. Make replaced elements `display: block` by default. (https://github.com/mozdevs/cssremedy/issues/14)
  2. Add `vertical-align: middle` to align replaced elements more sensibly by default. (https://github.com/jensimmons/cssremedy/issues/14#issuecomment-634934210)
    This can trigger a poorly considered lint error in some tools but is included by design.
  */
  "img,svg,video,canvas,audio,iframe,embed,object": {
    display: "block",
    /* 1 */
    verticalAlign: "middle"
  },
  /* 2 */
  /*
  Constrain images and videos to the parent width and preserve their intrinsic aspect ratio. (https://github.com/mozdevs/cssremedy/issues/14)
  */
  "img,video": {
    maxWidth: "100%",
    height: "auto"
  },
  /* Make elements with the HTML hidden attribute stay hidden by default */
  "[hidden]": {
    display: "none"
  }
}, yt = [
  /* arbitrary properties: [paint-order:markers] */
  c("\\[([-\\w]+):(.+)]", ({ 1: e, 2: t }, r) => ({
    "@layer overrides": {
      "&": {
        [e]: I(`[${t}]`, "", r)
      }
    }
  })),
  /* Styling based on parent and peer state */
  c("(group|peer)([~/][^-[]+)?", ({ input: e }, { h: t }) => [
    {
      c: t(e)
    }
  ]),
  /* LAYOUT */
  d("aspect-", "aspectRatio"),
  c("container", (e, { theme: t }) => {
    let { screens: r = t("screens"), center: o, padding: n } = t("container"), l = {
      width: "100%",
      marginRight: o && "auto",
      marginLeft: o && "auto",
      ...s("xs")
    };
    for (let a in r) {
      let i = r[a];
      typeof i == "string" && (l[ve(i)] = {
        "&": {
          maxWidth: i,
          ...s(a)
        }
      });
    }
    return l;
    function s(a) {
      let i = n && (typeof n == "string" ? n : n[a] || n.DEFAULT);
      if (i) return {
        paddingRight: i,
        paddingLeft: i
      };
    }
  }),
  // Content
  d("content-", "content", ({ _: e }) => ({
    "--tw-content": e,
    content: "var(--tw-content)"
  })),
  // Box Decoration Break
  c("(?:box-)?decoration-(slice|clone)", "boxDecorationBreak"),
  // Box Sizing
  c("box-(border|content)", "boxSizing", ({ 1: e }) => e + "-box"),
  // Display
  c("hidden", {
    display: "none"
  }),
  // Table Layout
  c("table-(auto|fixed)", "tableLayout"),
  c([
    "(block|flex|table|grid|inline|contents|flow-root|list-item)",
    "(inline-(block|flex|table|grid))",
    "(table-(caption|cell|column|row|(column|row|footer|header)-group))"
  ], "display"),
  // Floats
  "(float)-(left|right|none)",
  // Clear
  "(clear)-(left|right|none|both)",
  // Overflow
  "(overflow(?:-[xy])?)-(auto|hidden|clip|visible|scroll)",
  // Isolation
  "(isolation)-(auto)",
  // Isolation
  c("isolate", "isolation"),
  // Object Fit
  c("object-(contain|cover|fill|none|scale-down)", "objectFit"),
  // Object Position
  d("object-", "objectPosition"),
  c("object-(top|bottom|center|(left|right)(-(top|bottom))?)", "objectPosition", K),
  // Overscroll Behavior
  c("overscroll(-[xy])?-(auto|contain|none)", ({ 1: e = "", 2: t }) => ({
    ["overscroll-behavior" + e]: t
  })),
  // Position
  c("(static|fixed|absolute|relative|sticky)", "position"),
  // Top / Right / Bottom / Left
  d("-?inset(-[xy])?(?:$|-)", "inset", ({ 1: e, _: t }) => ({
    top: e != "-x" && t,
    right: e != "-y" && t,
    bottom: e != "-x" && t,
    left: e != "-y" && t
  })),
  d("-?(top|bottom|left|right)(?:$|-)", "inset"),
  // Visibility
  c("(visible|collapse)", "visibility"),
  c("invisible", {
    visibility: "hidden"
  }),
  // Z-Index
  d("-?z-", "zIndex"),
  /* FLEXBOX */
  // Flex Direction
  c("flex-((row|col)(-reverse)?)", "flexDirection", je),
  c("flex-(wrap|wrap-reverse|nowrap)", "flexWrap"),
  d("(flex-(?:grow|shrink))(?:$|-)"),
  /*, 'flex-grow' | flex-shrink */
  d("(flex)-"),
  /*, 'flex' */
  d("grow(?:$|-)", "flexGrow"),
  d("shrink(?:$|-)", "flexShrink"),
  d("basis-", "flexBasis"),
  d("-?(order)-"),
  /*, 'order' */
  "-?(order)-(\\d+)",
  /* GRID */
  // Grid Template Columns
  d("grid-cols-", "gridTemplateColumns"),
  c("grid-cols-(\\d+)", "gridTemplateColumns", Ve),
  // Grid Column Start / End
  d("col-", "gridColumn"),
  c("col-(span)-(\\d+)", "gridColumn", Me),
  d("col-start-", "gridColumnStart"),
  c("col-start-(auto|\\d+)", "gridColumnStart"),
  d("col-end-", "gridColumnEnd"),
  c("col-end-(auto|\\d+)", "gridColumnEnd"),
  // Grid Template Rows
  d("grid-rows-", "gridTemplateRows"),
  c("grid-rows-(\\d+)", "gridTemplateRows", Ve),
  // Grid Row Start / End
  d("row-", "gridRow"),
  c("row-(span)-(\\d+)", "gridRow", Me),
  d("row-start-", "gridRowStart"),
  c("row-start-(auto|\\d+)", "gridRowStart"),
  d("row-end-", "gridRowEnd"),
  c("row-end-(auto|\\d+)", "gridRowEnd"),
  // Grid Auto Flow
  c("grid-flow-((row|col)(-dense)?)", "gridAutoFlow", (e) => K(je(e))),
  c("grid-flow-(dense)", "gridAutoFlow"),
  // Grid Auto Columns
  d("auto-cols-", "gridAutoColumns"),
  // Grid Auto Rows
  d("auto-rows-", "gridAutoRows"),
  // Gap
  d("gap-x(?:$|-)", "gap", "columnGap"),
  d("gap-y(?:$|-)", "gap", "rowGap"),
  d("gap(?:$|-)", "gap"),
  /* BOX ALIGNMENT */
  // Justify Items
  // Justify Self
  "(justify-(?:items|self))-",
  // Justify Content
  c("justify-", "justifyContent", De),
  // Align Content
  // Align Items
  // Align Self
  c("(content|items|self)-", (e) => ({
    ["align-" + e[1]]: De(e)
  })),
  // Place Content
  // Place Items
  // Place Self
  c("(place-(content|items|self))-", ({ 1: e, $$: t }) => ({
    [e]: ("wun".includes(t[3]) ? "space-" : "") + t
  })),
  /* SPACING */
  // Padding
  d("p([xytrbl])?(?:$|-)", "padding", G("padding")),
  // Margin
  d("-?m([xytrbl])?(?:$|-)", "margin", G("margin")),
  // Space Between
  d("-?space-(x|y)(?:$|-)", "space", ({ 1: e, _: t }) => ({
    "&>:not([hidden])~:not([hidden])": {
      [`--tw-space-${e}-reverse`]: "0",
      ["margin-" + {
        y: "top",
        x: "left"
      }[e]]: `calc(${t} * calc(1 - var(--tw-space-${e}-reverse)))`,
      ["margin-" + {
        y: "bottom",
        x: "right"
      }[e]]: `calc(${t} * var(--tw-space-${e}-reverse))`
    }
  })),
  c("space-(x|y)-reverse", ({ 1: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      [`--tw-space-${e}-reverse`]: "1"
    }
  })),
  /* SIZING */
  // Width
  d("w-", "width"),
  // Min-Width
  d("min-w-", "minWidth"),
  // Max-Width
  d("max-w-", "maxWidth"),
  // Height
  d("h-", "height"),
  // Min-Height
  d("min-h-", "minHeight"),
  // Max-Height
  d("max-h-", "maxHeight"),
  /* TYPOGRAPHY */
  // Font Weight
  d("font-", "fontWeight"),
  // Font Family
  d("font-", "fontFamily", ({ _: e }) => typeof (e = w(e))[1] == "string" ? {
    fontFamily: W(e)
  } : {
    fontFamily: W(e[0]),
    ...e[1]
  }),
  // Font Smoothing
  c("antialiased", {
    WebkitFontSmoothing: "antialiased",
    MozOsxFontSmoothing: "grayscale"
  }),
  c("subpixel-antialiased", {
    WebkitFontSmoothing: "auto",
    MozOsxFontSmoothing: "auto"
  }),
  // Font Style
  c("italic", "fontStyle"),
  c("not-italic", {
    fontStyle: "normal"
  }),
  // Font Variant Numeric
  c("(ordinal|slashed-zero|(normal|lining|oldstyle|proportional|tabular)-nums|(diagonal|stacked)-fractions)", ({ 1: e, 2: t = "", 3: r }) => (
    // normal-nums
    t == "normal" ? {
      fontVariantNumeric: "normal"
    } : {
      ["--tw-" + (r ? (
        // diagonal-fractions, stacked-fractions
        "numeric-fraction"
      ) : "pt".includes(t[0]) ? (
        // proportional-nums, tabular-nums
        "numeric-spacing"
      ) : t ? (
        // lining-nums, oldstyle-nums
        "numeric-figure"
      ) : (
        // ordinal, slashed-zero
        e
      ))]: e,
      fontVariantNumeric: "var(--tw-ordinal) var(--tw-slashed-zero) var(--tw-numeric-figure) var(--tw-numeric-spacing) var(--tw-numeric-fraction)",
      ...L({
        "--tw-ordinal": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-slashed-zero": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-numeric-figure": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-numeric-spacing": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-numeric-fraction": "var(--tw-empty,/*!*/ /*!*/)"
      })
    }
  )),
  // Letter Spacing
  d("tracking-", "letterSpacing"),
  // Line Height
  d("leading-", "lineHeight"),
  // List Style Position
  c("list-(inside|outside)", "listStylePosition"),
  // List Style Type
  d("list-", "listStyleType"),
  c("list-", "listStyleType"),
  // Placeholder Opacity
  d("placeholder-opacity-", "placeholderOpacity", ({ _: e }) => ({
    "&::placeholder": {
      "--tw-placeholder-opacity": e
    }
  })),
  // Placeholder Color
  S("placeholder-", {
    property: "color",
    selector: "&::placeholder"
  }),
  // Text Alignment
  c("text-(left|center|right|justify|start|end)", "textAlign"),
  c("text-(ellipsis|clip)", "textOverflow"),
  // Text Opacity
  d("text-opacity-", "textOpacity", "--tw-text-opacity"),
  // Text Color
  S("text-", {
    property: "color"
  }),
  // Font Size
  d("text-", "fontSize", ({ _: e }) => typeof e == "string" ? {
    fontSize: e
  } : {
    fontSize: e[0],
    ...typeof e[1] == "string" ? {
      lineHeight: e[1]
    } : e[1]
  }),
  // Text Indent
  d("indent-", "textIndent"),
  // Text Decoration
  c("(overline|underline|line-through)", "textDecorationLine"),
  c("no-underline", {
    textDecorationLine: "none"
  }),
  // Text Underline offset
  d("underline-offset-", "textUnderlineOffset"),
  // Text Decoration Color
  S("decoration-", {
    section: "textDecorationColor",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Text Decoration Thickness
  d("decoration-", "textDecorationThickness"),
  // Text Decoration Style
  c("decoration-", "textDecorationStyle"),
  // Text Transform
  c("(uppercase|lowercase|capitalize)", "textTransform"),
  c("normal-case", {
    textTransform: "none"
  }),
  // Text Overflow
  c("truncate", {
    overflow: "hidden",
    whiteSpace: "nowrap",
    textOverflow: "ellipsis"
  }),
  // Vertical Alignment
  c("align-", "verticalAlign"),
  // Whitespace
  c("whitespace-", "whiteSpace"),
  // Word Break
  c("break-normal", {
    wordBreak: "normal",
    overflowWrap: "normal"
  }),
  c("break-words", {
    overflowWrap: "break-word"
  }),
  c("break-all", {
    wordBreak: "break-all"
  }),
  c("break-keep", {
    wordBreak: "keep-all"
  }),
  // Caret Color
  S("caret-", {
    // section: 'caretColor',
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Accent Color
  S("accent-", {
    // section: 'accentColor',
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Gradient Color Stops
  c("bg-gradient-to-([trbl]|[tb][rl])", "backgroundImage", ({ 1: e }) => `linear-gradient(to ${P(e, " ")},var(--tw-gradient-stops))`),
  S("from-", {
    section: "gradientColorStops",
    opacityVariable: !1,
    opacitySection: "opacity"
  }, ({ _: e }) => ({
    "--tw-gradient-from": e.value,
    "--tw-gradient-to": e.color({
      opacityValue: "0"
    }),
    "--tw-gradient-stops": "var(--tw-gradient-from),var(--tw-gradient-to)"
  })),
  S("via-", {
    section: "gradientColorStops",
    opacityVariable: !1,
    opacitySection: "opacity"
  }, ({ _: e }) => ({
    "--tw-gradient-to": e.color({
      opacityValue: "0"
    }),
    "--tw-gradient-stops": `var(--tw-gradient-from),${e.value},var(--tw-gradient-to)`
  })),
  S("to-", {
    section: "gradientColorStops",
    property: "--tw-gradient-to",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  /* BACKGROUNDS */
  // Background Attachment
  c("bg-(fixed|local|scroll)", "backgroundAttachment"),
  // Background Origin
  c("bg-origin-(border|padding|content)", "backgroundOrigin", ({ 1: e }) => e + "-box"),
  // Background Repeat
  c([
    "bg-(no-repeat|repeat(-[xy])?)",
    "bg-repeat-(round|space)"
  ], "backgroundRepeat"),
  // Background Blend Mode
  c("bg-blend-", "backgroundBlendMode"),
  // Background Clip
  c("bg-clip-(border|padding|content|text)", "backgroundClip", ({ 1: e }) => e + (e == "text" ? "" : "-box")),
  // Background Opacity
  d("bg-opacity-", "backgroundOpacity", "--tw-bg-opacity"),
  // Background Color
  // bg-${backgroundColor}/${backgroundOpacity}
  S("bg-", {
    section: "backgroundColor"
  }),
  // Background Image
  // supported arbitrary types are: length, color, angle, list
  d("bg-", "backgroundImage"),
  // Background Position
  d("bg-", "backgroundPosition"),
  c("bg-(top|bottom|center|(left|right)(-(top|bottom))?)", "backgroundPosition", K),
  // Background Size
  d("bg-", "backgroundSize"),
  /* BORDERS */
  // Border Radius
  d("rounded(?:$|-)", "borderRadius"),
  d("rounded-([trbl]|[tb][rl])(?:$|-)", "borderRadius", ({ 1: e, _: t }) => {
    let r = {
      t: [
        "tl",
        "tr"
      ],
      r: [
        "tr",
        "br"
      ],
      b: [
        "bl",
        "br"
      ],
      l: [
        "bl",
        "tl"
      ]
    }[e] || [
      e,
      e
    ];
    return {
      [`border-${P(r[0])}-radius`]: t,
      [`border-${P(r[1])}-radius`]: t
    };
  }),
  // Border Collapse
  c("border-(collapse|separate)", "borderCollapse"),
  // Border Opacity
  d("border-opacity(?:$|-)", "borderOpacity", "--tw-border-opacity"),
  // Border Style
  c("border-(solid|dashed|dotted|double|none)", "borderStyle"),
  // Border Spacing
  d("border-spacing(-[xy])?(?:$|-)", "borderSpacing", ({ 1: e, _: t }) => ({
    ...L({
      "--tw-border-spacing-x": "0",
      "--tw-border-spacing-y": "0"
    }),
    ["--tw-border-spacing" + (e || "-x")]: t,
    ["--tw-border-spacing" + (e || "-y")]: t,
    "border-spacing": "var(--tw-border-spacing-x) var(--tw-border-spacing-y)"
  })),
  // Border Color
  S("border-([xytrbl])-", {
    section: "borderColor"
  }, G("border", "Color")),
  S("border-"),
  // Border Width
  d("border-([xytrbl])(?:$|-)", "borderWidth", G("border", "Width")),
  d("border(?:$|-)", "borderWidth"),
  // Divide Opacity
  d("divide-opacity(?:$|-)", "divideOpacity", ({ _: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      "--tw-divide-opacity": e
    }
  })),
  // Divide Style
  c("divide-(solid|dashed|dotted|double|none)", ({ 1: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      borderStyle: e
    }
  })),
  // Divide Width
  c("divide-([xy]-reverse)", ({ 1: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      ["--tw-divide-" + e]: "1"
    }
  })),
  d("divide-([xy])(?:$|-)", "divideWidth", ({ 1: e, _: t }) => {
    let r = {
      x: "lr",
      y: "tb"
    }[e];
    return {
      "&>:not([hidden])~:not([hidden])": {
        [`--tw-divide-${e}-reverse`]: "0",
        [`border-${P(r[0])}Width`]: `calc(${t} * calc(1 - var(--tw-divide-${e}-reverse)))`,
        [`border-${P(r[1])}Width`]: `calc(${t} * var(--tw-divide-${e}-reverse))`
      }
    };
  }),
  // Divide Color
  S("divide-", {
    // section: $0.replace('-', 'Color') -> 'divideColor'
    property: "borderColor",
    // opacityVariable: '--tw-border-opacity',
    // opacitySection: section.replace('Color', 'Opacity') -> 'divideOpacity'
    selector: "&>:not([hidden])~:not([hidden])"
  }),
  // Ring Offset Opacity
  d("ring-opacity(?:$|-)", "ringOpacity", "--tw-ring-opacity"),
  // Ring Offset Color
  S("ring-offset-", {
    // section: 'ringOffsetColor',
    property: "--tw-ring-offset-color",
    opacityVariable: !1
  }),
  // opacitySection: section.replace('Color', 'Opacity') -> 'ringOffsetOpacity'
  // Ring Offset Width
  d("ring-offset(?:$|-)", "ringOffsetWidth", "--tw-ring-offset-width"),
  // Ring Inset
  c("ring-inset", {
    "--tw-ring-inset": "inset"
  }),
  // Ring Color
  S("ring-", {
    // section: 'ringColor',
    property: "--tw-ring-color"
  }),
  // opacityVariable: '--tw-ring-opacity',
  // opacitySection: section.replace('Color', 'Opacity') -> 'ringOpacity'
  // Ring Width
  d("ring(?:$|-)", "ringWidth", ({ _: e }, { theme: t }) => ({
    ...L({
      "--tw-ring-offset-shadow": "0 0 #0000",
      "--tw-ring-shadow": "0 0 #0000",
      "--tw-shadow": "0 0 #0000",
      "--tw-shadow-colored": "0 0 #0000",
      // Within own declaration to have the defaults above to be merged with defaults from shadow
      "&": {
        "--tw-ring-inset": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-ring-offset-width": t("ringOffsetWidth", "", "0px"),
        "--tw-ring-offset-color": U(t("ringOffsetColor", "", "#fff")),
        "--tw-ring-color": U(t("ringColor", "", "#93c5fd"), {
          opacityVariable: "--tw-ring-opacity"
        }),
        "--tw-ring-opacity": t("ringOpacity", "", "0.5")
      }
    }),
    "--tw-ring-offset-shadow": "var(--tw-ring-inset) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)",
    "--tw-ring-shadow": `var(--tw-ring-inset) 0 0 0 calc(${e} + var(--tw-ring-offset-width)) var(--tw-ring-color)`,
    boxShadow: "var(--tw-ring-offset-shadow),var(--tw-ring-shadow),var(--tw-shadow)"
  })),
  /* EFFECTS */
  // Box Shadow Color
  S("shadow-", {
    section: "boxShadowColor",
    opacityVariable: !1,
    opacitySection: "opacity"
  }, ({ _: e }) => ({
    "--tw-shadow-color": e.value,
    "--tw-shadow": "var(--tw-shadow-colored)"
  })),
  // Box Shadow
  d("shadow(?:$|-)", "boxShadow", ({ _: e }) => ({
    ...L({
      "--tw-ring-offset-shadow": "0 0 #0000",
      "--tw-ring-shadow": "0 0 #0000",
      "--tw-shadow": "0 0 #0000",
      "--tw-shadow-colored": "0 0 #0000"
    }),
    "--tw-shadow": W(e),
    // replace all colors with reference to --tw-shadow-colored
    // this matches colors after non-comma char (keyword, offset) before comma or the end
    "--tw-shadow-colored": W(e).replace(/([^,]\s+)(?:#[a-f\d]+|(?:(?:hsl|rgb)a?|hwb|lab|lch|color|var)\(.+?\)|[a-z]+)(,|$)/g, "$1var(--tw-shadow-color)$2"),
    boxShadow: "var(--tw-ring-offset-shadow),var(--tw-ring-shadow),var(--tw-shadow)"
  })),
  // Opacity
  d("(opacity)-"),
  /*, 'opacity' */
  // Mix Blend Mode
  c("mix-blend-", "mixBlendMode"),
  /* FILTERS */
  ...We(),
  ...We("backdrop-"),
  /* TRANSITIONS AND ANIMATION */
  // Transition Property
  d("transition(?:$|-)", "transitionProperty", (e, { theme: t }) => ({
    transitionProperty: W(e),
    transitionTimingFunction: e._ == "none" ? void 0 : W(t("transitionTimingFunction", "")),
    transitionDuration: e._ == "none" ? void 0 : W(t("transitionDuration", ""))
  })),
  // Transition Duration
  d("duration(?:$|-)", "transitionDuration", "transitionDuration", W),
  // Transition Timing Function
  d("ease(?:$|-)", "transitionTimingFunction", "transitionTimingFunction", W),
  // Transition Delay
  d("delay(?:$|-)", "transitionDelay", "transitionDelay", W),
  d("animate(?:$|-)", "animation", (e, { theme: t, h: r, e: o }) => {
    let n = W(e), l = n.split(" "), s = t("keyframes", l[0]);
    return s ? {
      ["@keyframes " + (l[0] = o(r(l[0])))]: s,
      animation: l.join(" ")
    } : {
      animation: n
    };
  }),
  /* TRANSFORMS */
  // Transform
  "(transform)-(none)",
  c("transform", ye),
  c("transform-(cpu|gpu)", ({ 1: e }) => ({
    "--tw-transform": _e(e == "gpu")
  })),
  // Scale
  d("scale(-[xy])?-", "scale", ({ 1: e, _: t }) => ({
    ["--tw-scale" + (e || "-x")]: t,
    ["--tw-scale" + (e || "-y")]: t,
    ...ye()
  })),
  // Rotate
  d("-?(rotate)-", "rotate", pe),
  // Translate
  d("-?(translate-[xy])-", "translate", pe),
  // Skew
  d("-?(skew-[xy])-", "skew", pe),
  // Transform Origin
  c("origin-(center|((top|bottom)(-(left|right))?)|left|right)", "transformOrigin", K),
  /* INTERACTIVITY */
  // Appearance
  "(appearance)-",
  // Columns
  d("(columns)-"),
  /*, 'columns' */
  "(columns)-(\\d+)",
  // Break Before, After and Inside
  "(break-(?:before|after|inside))-",
  // Cursor
  d("(cursor)-"),
  /*, 'cursor' */
  "(cursor)-",
  // Scroll Snap Type
  c("snap-(none)", "scroll-snap-type"),
  c("snap-(x|y|both)", ({ 1: e }) => ({
    ...L({
      "--tw-scroll-snap-strictness": "proximity"
    }),
    "scroll-snap-type": e + " var(--tw-scroll-snap-strictness)"
  })),
  c("snap-(mandatory|proximity)", "--tw-scroll-snap-strictness"),
  // Scroll Snap Align
  c("snap-(?:(start|end|center)|align-(none))", "scroll-snap-align"),
  // Scroll Snap Stop
  c("snap-(normal|always)", "scroll-snap-stop"),
  c("scroll-(auto|smooth)", "scroll-behavior"),
  // Scroll Margin
  // Padding
  d("scroll-p([xytrbl])?(?:$|-)", "padding", G("scroll-padding")),
  // Margin
  d("-?scroll-m([xytrbl])?(?:$|-)", "scroll-margin", G("scroll-margin")),
  // Touch Action
  c("touch-(auto|none|manipulation)", "touch-action"),
  c("touch-(pinch-zoom|pan-(?:(x|left|right)|(y|up|down)))", ({ 1: e, 2: t, 3: r }) => ({
    ...L({
      "--tw-pan-x": "var(--tw-empty,/*!*/ /*!*/)",
      "--tw-pan-y": "var(--tw-empty,/*!*/ /*!*/)",
      "--tw-pinch-zoom": "var(--tw-empty,/*!*/ /*!*/)",
      "--tw-touch-action": "var(--tw-pan-x) var(--tw-pan-y) var(--tw-pinch-zoom)"
    }),
    // x, left, right -> pan-x
    // y, up, down -> pan-y
    // -> pinch-zoom
    [`--tw-${t ? "pan-x" : r ? "pan-y" : e}`]: e,
    "touch-action": "var(--tw-touch-action)"
  })),
  // Outline Style
  c("outline-none", {
    outline: "2px solid transparent",
    "outline-offset": "2px"
  }),
  c("outline", {
    outlineStyle: "solid"
  }),
  c("outline-(dashed|dotted|double)", "outlineStyle"),
  // Outline Offset
  d("-?(outline-offset)-"),
  /*, 'outlineOffset'*/
  // Outline Color
  S("outline-", {
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Outline Width
  d("outline-", "outlineWidth"),
  // Pointer Events
  "(pointer-events)-",
  // Will Change
  d("(will-change)-"),
  /*, 'willChange' */
  "(will-change)-",
  // Resize
  [
    "resize(?:-(none|x|y))?",
    "resize",
    ({ 1: e }) => ({
      x: "horizontal",
      y: "vertical"
    })[e] || e || "both"
  ],
  // User Select
  c("select-(none|text|all|auto)", "userSelect"),
  /* SVG */
  // Fill, Stroke
  S("fill-", {
    section: "fill",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  S("stroke-", {
    section: "stroke",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Stroke Width
  d("stroke-", "strokeWidth"),
  /* ACCESSIBILITY */
  // Screen Readers
  c("sr-only", {
    position: "absolute",
    width: "1px",
    height: "1px",
    padding: "0",
    margin: "-1px",
    overflow: "hidden",
    whiteSpace: "nowrap",
    clip: "rect(0,0,0,0)",
    borderWidth: "0"
  }),
  c("not-sr-only", {
    position: "static",
    width: "auto",
    height: "auto",
    padding: "0",
    margin: "0",
    overflow: "visible",
    whiteSpace: "normal",
    clip: "auto"
  })
];
function K(e) {
  return (typeof e == "string" ? e : e[1]).replace(/-/g, " ").trim();
}
function je(e) {
  return (typeof e == "string" ? e : e[1]).replace("col", "column");
}
function P(e, t = "-") {
  let r = [];
  for (let o of e) r.push({
    t: "top",
    r: "right",
    b: "bottom",
    l: "left"
  }[o]);
  return r.join(t);
}
function W(e) {
  return e && "" + (e._ || e);
}
function De({ $$: e }) {
  return ({
    // /* aut*/ o: '',
    /* sta*/
    r: (
      /*t*/
      "flex-"
    ),
    /* end*/
    "": "flex-",
    // /* cen*/ t /*er*/: '',
    /* bet*/
    w: (
      /*een*/
      "space-"
    ),
    /* aro*/
    u: (
      /*nd*/
      "space-"
    ),
    /* eve*/
    n: (
      /*ly*/
      "space-"
    )
  }[e[3] || ""] || "") + e;
}
function G(e, t = "") {
  return ({ 1: r, _: o }) => {
    let n = {
      x: "lr",
      y: "tb"
    }[r] || r + r;
    return n ? {
      ...ee(e + "-" + P(n[0]) + t, o),
      ...ee(e + "-" + P(n[1]) + t, o)
    } : ee(e + t, o);
  };
}
function We(e = "") {
  let t = [
    "blur",
    "brightness",
    "contrast",
    "grayscale",
    "hue-rotate",
    "invert",
    e && "opacity",
    "saturate",
    "sepia",
    !e && "drop-shadow"
  ].filter(Boolean), r = {};
  for (let o of t) r[`--tw-${e}${o}`] = "var(--tw-empty,/*!*/ /*!*/)";
  return r = {
    // move defaults
    ...L(r),
    // add default filter which allows standalone usage
    [`${e}filter`]: t.map((o) => `var(--tw-${e}${o})`).join(" ")
  }, [
    `(${e}filter)-(none)`,
    c(`${e}filter`, r),
    ...t.map((o) => d(
      // hue-rotate can be negated
      `${o[0] == "h" ? "-?" : ""}(${e}${o})(?:$|-)`,
      o,
      ({ 1: n, _: l }) => ({
        [`--tw-${n}`]: w(l).map((s) => `${o}(${s})`).join(" "),
        ...r
      })
    ))
  ];
}
function pe({ 1: e, _: t }) {
  return {
    ["--tw-" + e]: t,
    ...ye()
  };
}
function ye() {
  return {
    ...L({
      "--tw-translate-x": "0",
      "--tw-translate-y": "0",
      "--tw-rotate": "0",
      "--tw-skew-x": "0",
      "--tw-skew-y": "0",
      "--tw-scale-x": "1",
      "--tw-scale-y": "1",
      "--tw-transform": _e()
    }),
    transform: "var(--tw-transform)"
  };
}
function _e(e) {
  return [
    e ? (
      // -gpu
      "translate3d(var(--tw-translate-x),var(--tw-translate-y),0)"
    ) : "translateX(var(--tw-translate-x)) translateY(var(--tw-translate-y))",
    "rotate(var(--tw-rotate))",
    "skewX(var(--tw-skew-x))",
    "skewY(var(--tw-skew-y))",
    "scaleX(var(--tw-scale-x))",
    "scaleY(var(--tw-scale-y))"
  ].join(" ");
}
function Me({ 1: e, 2: t }) {
  return `${e} ${t} / ${e} ${t}`;
}
function Ve({ 1: e }) {
  return `repeat(${e},minmax(0,1fr))`;
}
function L(e) {
  return {
    "@layer defaults": {
      "*,::before,::after": e,
      "::backdrop": e
    }
  };
}
let xt = [
  [
    "sticky",
    "@supports ((position: -webkit-sticky) or (position:sticky))"
  ],
  [
    "motion-reduce",
    "@media (prefers-reduced-motion:reduce)"
  ],
  [
    "motion-safe",
    "@media (prefers-reduced-motion:no-preference)"
  ],
  [
    "print",
    "@media print"
  ],
  [
    "(portrait|landscape)",
    ({ 1: e }) => `@media (orientation:${e})`
  ],
  [
    "contrast-(more|less)",
    ({ 1: e }) => `@media (prefers-contrast:${e})`
  ],
  [
    "(first-(letter|line)|placeholder|backdrop|before|after)",
    ({ 1: e }) => `&::${e}`
  ],
  [
    "(marker|selection)",
    ({ 1: e }) => `& *::${e},&::${e}`
  ],
  [
    "file",
    "&::file-selector-button"
  ],
  [
    "(first|last|only)",
    ({ 1: e }) => `&:${e}-child`
  ],
  [
    "even",
    "&:nth-child(2n)"
  ],
  [
    "odd",
    "&:nth-child(odd)"
  ],
  [
    "open",
    "&[open]"
  ],
  // All other pseudo classes are already supported by twind
  [
    "(aria|data)-",
    ({
      1: e,
      /* aria or data */
      $$: t
    }, r) => t && `&[${e}-${// aria-asc or data-checked -> from theme
    r.theme(e, t) || // aria-[...] or data-[...]
    I(t, "", r) || // default handling
    `${t}="true"`}]`
  ],
  /* Styling based on parent and peer state */
  // Groups classes like: group-focus and group-hover
  // these need to add a marker selector with the pseudo class
  // => '.group:focus .group-focus:selector'
  [
    "((group|peer)(~[^-[]+)?)(-\\[(.+)]|[-[].+?)(\\/.+)?",
    ({ 2: e, 3: t = "", 4: r, 5: o = "", 6: n = t }, { e: l, h: s, v: a }) => {
      let i = Q(o) || (r[0] == "[" ? r : a(r.slice(1)));
      return `${(i.includes("&") ? i : "&" + i).replace(/&/g, `:merge(.${l(s(e + n))})`)}${e[0] == "p" ? "~" : " "}&`;
    }
  ],
  // direction variants
  [
    "(ltr|rtl)",
    ({ 1: e }) => `[dir="${e}"] &`
  ],
  [
    "supports-",
    ({ $$: e }, t) => {
      if (e && (e = t.theme("supports", e) || I(e, "", t)), e) return e.includes(":") || (e += ":var(--tw)"), /^\w*\s*\(/.test(e) || (e = `(${e})`), // Chrome has a bug where `(condtion1)or(condition2)` is not valid
      // But `(condition1) or (condition2)` is supported.
      `@supports ${e.replace(/\b(and|or|not)\b/g, " $1 ").trim()}`;
    }
  ],
  [
    "max-",
    ({ $$: e }, t) => {
      if (e && (e = t.theme("screens", e) || I(e, "", t)), typeof e == "string") return `@media not all and (min-width:${e})`;
    }
  ],
  [
    "min-",
    ({ $$: e }, t) => (e && (e = I(e, "", t)), e && `@media (min-width:${e})`)
  ],
  // Arbitrary variants
  [
    /^\[(.+)]$/,
    ({ 1: e }) => /[&@]/.test(e) && Q(e).replace(/[}]+$/, "").split("{")
  ]
];
function vt({ colors: e, disablePreflight: t } = {}) {
  return {
    // allow other preflight to run
    preflight: t ? void 0 : wt,
    theme: {
      ...we,
      colors: {
        inherit: "inherit",
        current: "currentColor",
        transparent: "transparent",
        black: "#000",
        white: "#fff",
        ...e
      }
    },
    variants: xt,
    rules: yt,
    finalize(r) {
      return (
        // automatically add `content: ''` to before and after so you don’t have to specify it unless you want a different value
        // ignore global, preflight, and auto added rules
        r.n && // only if there are declarations
        r.d && // and it has a ::before or ::after selector
        r.r.some((o) => /^&::(before|after)$/.test(o)) && // there is no content property yet
        !/(^|;)content:/.test(r.d) ? {
          ...r,
          d: "content:var(--tw-content);" + r.d
        } : r
      );
    }
  };
}
let kt = {
  50: "#f8fafc",
  100: "#f1f5f9",
  200: "#e2e8f0",
  300: "#cbd5e1",
  400: "#94a3b8",
  500: "#64748b",
  600: "#475569",
  700: "#334155",
  800: "#1e293b",
  900: "#0f172a"
}, St = {
  50: "#f9fafb",
  100: "#f3f4f6",
  200: "#e5e7eb",
  300: "#d1d5db",
  400: "#9ca3af",
  500: "#6b7280",
  600: "#4b5563",
  700: "#374151",
  800: "#1f2937",
  900: "#111827"
}, Ct = {
  50: "#fafafa",
  100: "#f4f4f5",
  200: "#e4e4e7",
  300: "#d4d4d8",
  400: "#a1a1aa",
  500: "#71717a",
  600: "#52525b",
  700: "#3f3f46",
  800: "#27272a",
  900: "#18181b"
}, $t = {
  50: "#fafafa",
  100: "#f5f5f5",
  200: "#e5e5e5",
  300: "#d4d4d4",
  400: "#a3a3a3",
  500: "#737373",
  600: "#525252",
  700: "#404040",
  800: "#262626",
  900: "#171717"
}, Rt = {
  50: "#fafaf9",
  100: "#f5f5f4",
  200: "#e7e5e4",
  300: "#d6d3d1",
  400: "#a8a29e",
  500: "#78716c",
  600: "#57534e",
  700: "#44403c",
  800: "#292524",
  900: "#1c1917"
}, Tt = {
  50: "#fef2f2",
  100: "#fee2e2",
  200: "#fecaca",
  300: "#fca5a5",
  400: "#f87171",
  500: "#ef4444",
  600: "#dc2626",
  700: "#b91c1c",
  800: "#991b1b",
  900: "#7f1d1d"
}, zt = {
  50: "#fff7ed",
  100: "#ffedd5",
  200: "#fed7aa",
  300: "#fdba74",
  400: "#fb923c",
  500: "#f97316",
  600: "#ea580c",
  700: "#c2410c",
  800: "#9a3412",
  900: "#7c2d12"
}, At = {
  50: "#fffbeb",
  100: "#fef3c7",
  200: "#fde68a",
  300: "#fcd34d",
  400: "#fbbf24",
  500: "#f59e0b",
  600: "#d97706",
  700: "#b45309",
  800: "#92400e",
  900: "#78350f"
}, Ft = {
  50: "#fefce8",
  100: "#fef9c3",
  200: "#fef08a",
  300: "#fde047",
  400: "#facc15",
  500: "#eab308",
  600: "#ca8a04",
  700: "#a16207",
  800: "#854d0e",
  900: "#713f12"
}, Et = {
  50: "#f7fee7",
  100: "#ecfccb",
  200: "#d9f99d",
  300: "#bef264",
  400: "#a3e635",
  500: "#84cc16",
  600: "#65a30d",
  700: "#4d7c0f",
  800: "#3f6212",
  900: "#365314"
}, Ot = {
  50: "#f0fdf4",
  100: "#dcfce7",
  200: "#bbf7d0",
  300: "#86efac",
  400: "#4ade80",
  500: "#22c55e",
  600: "#16a34a",
  700: "#15803d",
  800: "#166534",
  900: "#14532d"
}, jt = {
  50: "#ecfdf5",
  100: "#d1fae5",
  200: "#a7f3d0",
  300: "#6ee7b7",
  400: "#34d399",
  500: "#10b981",
  600: "#059669",
  700: "#047857",
  800: "#065f46",
  900: "#064e3b"
}, Dt = {
  50: "#f0fdfa",
  100: "#ccfbf1",
  200: "#99f6e4",
  300: "#5eead4",
  400: "#2dd4bf",
  500: "#14b8a6",
  600: "#0d9488",
  700: "#0f766e",
  800: "#115e59",
  900: "#134e4a"
}, Wt = {
  50: "#ecfeff",
  100: "#cffafe",
  200: "#a5f3fc",
  300: "#67e8f9",
  400: "#22d3ee",
  500: "#06b6d4",
  600: "#0891b2",
  700: "#0e7490",
  800: "#155e75",
  900: "#164e63"
}, Mt = {
  50: "#f0f9ff",
  100: "#e0f2fe",
  200: "#bae6fd",
  300: "#7dd3fc",
  400: "#38bdf8",
  500: "#0ea5e9",
  600: "#0284c7",
  700: "#0369a1",
  800: "#075985",
  900: "#0c4a6e"
}, Vt = {
  50: "#eff6ff",
  100: "#dbeafe",
  200: "#bfdbfe",
  300: "#93c5fd",
  400: "#60a5fa",
  500: "#3b82f6",
  600: "#2563eb",
  700: "#1d4ed8",
  800: "#1e40af",
  900: "#1e3a8a"
}, Lt = {
  50: "#eef2ff",
  100: "#e0e7ff",
  200: "#c7d2fe",
  300: "#a5b4fc",
  400: "#818cf8",
  500: "#6366f1",
  600: "#4f46e5",
  700: "#4338ca",
  800: "#3730a3",
  900: "#312e81"
}, Ut = {
  50: "#f5f3ff",
  100: "#ede9fe",
  200: "#ddd6fe",
  300: "#c4b5fd",
  400: "#a78bfa",
  500: "#8b5cf6",
  600: "#7c3aed",
  700: "#6d28d9",
  800: "#5b21b6",
  900: "#4c1d95"
}, It = {
  50: "#faf5ff",
  100: "#f3e8ff",
  200: "#e9d5ff",
  300: "#d8b4fe",
  400: "#c084fc",
  500: "#a855f7",
  600: "#9333ea",
  700: "#7e22ce",
  800: "#6b21a8",
  900: "#581c87"
}, Bt = {
  50: "#fdf4ff",
  100: "#fae8ff",
  200: "#f5d0fe",
  300: "#f0abfc",
  400: "#e879f9",
  500: "#d946ef",
  600: "#c026d3",
  700: "#a21caf",
  800: "#86198f",
  900: "#701a75"
}, Ht = {
  50: "#fdf2f8",
  100: "#fce7f3",
  200: "#fbcfe8",
  300: "#f9a8d4",
  400: "#f472b6",
  500: "#ec4899",
  600: "#db2777",
  700: "#be185d",
  800: "#9d174d",
  900: "#831843"
}, Nt = {
  50: "#fff1f2",
  100: "#ffe4e6",
  200: "#fecdd3",
  300: "#fda4af",
  400: "#fb7185",
  500: "#f43f5e",
  600: "#e11d48",
  700: "#be123c",
  800: "#9f1239",
  900: "#881337"
}, Pt = {
  __proto__: null,
  slate: kt,
  gray: St,
  zinc: Ct,
  neutral: $t,
  stone: Rt,
  red: Tt,
  orange: zt,
  amber: At,
  yellow: Ft,
  lime: Et,
  green: Ot,
  emerald: jt,
  teal: Dt,
  cyan: Wt,
  sky: Mt,
  blue: Vt,
  indigo: Lt,
  violet: Ut,
  purple: It,
  fuchsia: Bt,
  pink: Ht,
  rose: Nt
};
function qt({ disablePreflight: e } = {}) {
  return vt({
    colors: Pt,
    disablePreflight: e
  });
}
const Gt = Ge({
  presets: [ht(), qt()]
});
(function() {
  const e = dt(Gt);
  class t extends e(HTMLElement) {
    constructor() {
      super(), this.attachShadow({ mode: "open" }), this.shadowRoot && (this.shadowRoot.innerHTML = this.innerHTML, this.innerHTML = "");
    }
  }
  customElements.define("twind-scope", t);
})();

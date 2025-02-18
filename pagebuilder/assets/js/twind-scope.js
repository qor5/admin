function Ut(e) {
  return [
    ...e.v,
    (e.i ? "!" : "") + e.n
  ].join(":");
}
function hi(e, t = ",") {
  return e.map(Ut).join(t);
}
let Ur = typeof CSS < "u" && CSS.escape || // Simplified: escaping only special characters
// Needed for NodeJS and Edge <79 (https://caniuse.com/mdn-api_css_escape)
((e) => e.replace(/[!"'`*+.,;:\\/<=>?@#$%&^|~()[\]{}]/g, "\\$&").replace(/^\d/, "\\3$& "));
function Ke(e) {
  for (var t = 9, r = e.length; r--; ) t = Math.imul(t ^ e.charCodeAt(r), 1597334677);
  return "#" + ((t ^ t >>> 9) >>> 0).toString(36);
}
function Ht(e, t = "@media ") {
  return t + E(e).map((r) => (typeof r == "string" && (r = {
    min: r
  }), r.raw || Object.keys(r).map((n) => `(${n}-width:${r[n]})`).join(" and "))).join(",");
}
function E(e = []) {
  return Array.isArray(e) ? e : e == null ? [] : [
    e
  ];
}
function yr(e) {
  return e;
}
function qt() {
}
let M = {
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
function Hr(e) {
  var t;
  return ((t = e.match(/[-=:;]/g)) == null ? void 0 : t.length) || 0;
}
function wt(e) {
  return Math.min(/(?:^|width[^\d]+)(\d+(?:.\d+)?)(p)?/.test(e) ? Math.max(0, 29.63 * (+RegExp.$1 / (RegExp.$2 ? 15 : 1)) ** 0.137 - 43) : 0, 15) << 22 | /* Shifts.responsive */
  Math.min(Hr(e), 15) << 18;
}
let gi = [
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
function Kt({ n: e, i: t, v: r = [] }, n, i, o) {
  e && (e = Ut({
    n: e,
    i: t,
    v: r
  })), o = [
    ...E(o)
  ];
  for (let s of r) {
    let l = n.theme("screens", s);
    for (let c of E(l && Ht(l) || n.v(s))) {
      var a;
      o.push(c), i |= l ? 67108864 | /* Shifts.screens */
      wt(c) : s == "dark" ? 1073741824 : (
        /* Shifts.darkMode */
        c[0] == "@" ? wt(c) : (a = c, // use first found pseudo-class
        1 << ~(/:([a-z-]+)/.test(a) && ~gi.indexOf(RegExp.$1.slice(2, 7)) || -18))
      );
    }
  }
  return {
    n: e,
    p: i,
    r: o,
    i: t
  };
}
let qr = /* @__PURE__ */ new Map();
function _t(e) {
  if (e.d) {
    let t = [], r = lt(
      // merge all conditions into a selector string
      e.r.reduce((n, i) => i[0] == "@" ? (t.push(i), n) : (
        // Go over the selector and replace the matching multiple selectors if any
        i ? lt(n, (o) => lt(
          i,
          // If the current condition has a nested selector replace it
          (a) => {
            let s = /(:merge\(.+?\))(:[a-z-]+|\\[.+])/.exec(a);
            if (s) {
              let l = o.indexOf(s[1]);
              return ~l ? (
                // [':merge(.group):hover .rule', ':merge(.group):focus &'] -> ':merge(.group):focus:hover .rule'
                // ':merge(.group)' + ':focus' + ':hover .rule'
                o.slice(0, l) + s[0] + o.slice(l + s[1].length)
              ) : (
                // [':merge(.peer):focus~&', ':merge(.group):hover &'] -> ':merge(.peer):focus~:merge(.group):hover &'
                ct(o, a)
              );
            }
            return ct(a, o);
          }
        )) : n
      ), "&"),
      // replace '&' with rule name or an empty string
      (n) => ct(n, e.n ? "." + Ur(e.n) : "")
    );
    return r && t.push(r.replace(/:merge\((.+?)\)/g, "$1")), t.reduceRight((n, i) => i + "{" + n + "}", e.d);
  }
}
function lt(e, t) {
  return e.replace(/ *((?:\(.+?\)|\[.+?\]|[^,])+) *(,|$)/g, (r, n, i) => t(n) + i);
}
function ct(e, t) {
  return e.replace(/&/g, t);
}
let xr = new Intl.Collator("en", {
  numeric: !0
});
function Kr(e, t) {
  for (var r = 0, n = e.length; r < n; ) {
    let i = n + r >> 1;
    0 >= Jr(e[i], t) ? r = i + 1 : n = i;
  }
  return n;
}
function Jr(e, t) {
  let r = e.p & M.o;
  return r == (t.p & M.o) && (r == M.b || r == M.o) ? 0 : e.p - t.p || e.o - t.o || xr.compare(wr(e.n), wr(t.n)) || xr.compare(_r(e.n), _r(t.n));
}
function wr(e) {
  return (e || "").split(/:/).pop().split("/").pop() || "\0";
}
function _r(e) {
  return (e || "").replace(/\W/g, (t) => String.fromCharCode(127 + t.charCodeAt(0))) + "\0";
}
function ft(e, t) {
  return Math.round(parseInt(e, 16) * t);
}
function Z(e, t = {}) {
  if (typeof e == "function") return e(t);
  let { opacityValue: r = "1", opacityVariable: n } = t, i = n ? `var(${n})` : r;
  if (e.includes("<alpha-value>")) return e.replace("<alpha-value>", i);
  if (e[0] == "#" && (e.length == 4 || e.length == 7)) {
    let o = (e.length - 1) / 3, a = [
      17,
      1,
      0.062272
    ][o - 1];
    return `rgba(${[
      ft(e.substr(1, o), a),
      ft(e.substr(1 + o, o), a),
      ft(e.substr(1 + 2 * o, o), a),
      i
    ]})`;
  }
  return i == "1" ? e : i == "0" ? "#0000" : (
    // convert rgb and hsl to alpha variant
    e.replace(/^(rgb|hsl)(\([^)]+)\)$/, `$1a$2,${i})`)
  );
}
function Gr(e, t, r, n, i = []) {
  return function o(a, { n: s, p: l, r: c = [], i: f }, u) {
    let h = [], b = "", _ = 0, C = 0;
    for (let m in a || {}) {
      var g, w;
      let y = a[m];
      if (m[0] == "@") {
        if (!y) continue;
        if (m[1] == "a") {
          h.push(...Gt(s, l, Ge("" + y), u, l, c, f, !0));
          continue;
        }
        if (m[1] == "l") {
          for (let S of E(y)) h.push(...o(S, {
            n: s,
            p: (g = M[m[7]], // Set layer (first reset, than set)
            l & -939524097 | g),
            r: m[7] == "d" ? [] : c,
            i: f
          }, u));
          continue;
        }
        if (m[1] == "i") {
          h.push(...E(y).map((S) => ({
            // before all layers
            p: -1,
            o: 0,
            r: [],
            d: m + " " + S
          })));
          continue;
        }
        if (m[1] == "k") {
          h.push({
            p: M.d,
            o: 0,
            r: [
              m
            ],
            d: o(y, {
              p: M.d
            }, u).map(_t).join("")
          });
          continue;
        }
        if (m[1] == "f") {
          h.push(...E(y).map((S) => ({
            p: M.d,
            o: 0,
            r: [
              m
            ],
            d: o(S, {
              p: M.d
            }, u).map(_t).join("")
          })));
          continue;
        }
      }
      if (typeof y != "object" || Array.isArray(y))
        m == "label" && y ? s = y + Ke(JSON.stringify([
          l,
          f,
          a
        ])) : (y || y === 0) && (m = m.replace(/[A-Z]/g, (S) => "-" + S.toLowerCase()), C += 1, _ = Math.max(_, (w = m)[0] == "-" ? 0 : Hr(w) + (/^(?:(border-(?!w|c|sty)|[tlbr].{2,4}m?$|c.{7,8}$)|([fl].{5}l|g.{8}$|pl))/.test(w) ? +!!RegExp.$1 || /* +1 */
        -!!RegExp.$2 : (
          /* -1 */
          0
        )) + 1), b += (b ? ";" : "") + E(y).map((S) => u.s(
          m,
          // support theme(...) function in values
          // calc(100vh - theme('spacing.12'))
          Jt("" + S, u.theme) + (f ? " !important" : "")
        )).join(";"));
      else if (m[0] == "@" || m.includes("&")) {
        let S = l;
        m[0] == "@" && (m = m.replace(/\bscreen\(([^)]+)\)/g, (q, P) => {
          let F = u.theme("screens", P);
          return F ? (S |= 67108864, /* Shifts.screens */
          Ht(F, "")) : q;
        }), S |= wt(m)), h.push(...o(y, {
          n: s,
          p: S,
          r: [
            ...c,
            m
          ],
          i: f
        }, u));
      } else
        h.push(...o(y, {
          p: l,
          r: [
            ...c,
            m
          ]
        }, u));
    }
    return (
      // PERF: prevent unshift using `rules = [{}]` above and then `rules[0] = {...}`
      h.unshift({
        n: s,
        p: l,
        o: (
          // number of declarations (descending)
          Math.max(0, 15 - C) + // greatest precedence of properties
          // if there is no property precedence this is most likely a custom property only declaration
          // these have the highest precedence
          1.5 * Math.min(_ || 15, 15)
        ),
        r: c,
        // stringified declarations
        d: b
      }), h.sort(Jr)
    );
  }(e, Kt(t, r, n, i), r);
}
function Jt(e, t) {
  return e.replace(/theme\((["'`])?(.+?)\1(?:\s*,\s*(["'`])?(.+?)\3)?\)/g, (r, n, i, o, a = "") => {
    let s = t(i, a);
    return typeof s == "function" && /color|fill|stroke/i.test(i) ? Z(s) : "" + E(s).filter((l) => Object(l) !== l);
  });
}
function Yr(e, t) {
  let r, n = [];
  for (let i of e)
    i.d && i.n ? (r == null ? void 0 : r.p) == i.p && "" + r.r == "" + i.r ? (r.c = [
      r.c,
      i.c
    ].filter(Boolean).join(" "), r.d = r.d + ";" + i.d) : n.push(r = {
      ...i,
      n: i.n && t
    }) : n.push({
      ...i,
      n: i.n && t
    });
  return n;
}
function Je(e, t, r = M.u, n, i) {
  let o = [];
  for (let a of e) for (let s of function(l, c, f, u, h) {
    l = {
      ...l,
      i: l.i || h
    };
    let b = function(_, C) {
      let g = qr.get(_.n);
      return g ? g(_, C) : C.r(_.n, _.v[0] == "dark");
    }(l, c);
    return b ? (
      // a list of class names
      typeof b == "string" ? ({ r: u, p: f } = Kt(l, c, f, u), Yr(Je(Ge(b), c, f, u, l.i), l.n)) : Array.isArray(b) ? b.map((_) => {
        var C, g;
        return {
          o: 0,
          ..._,
          r: [
            ...E(u),
            ...E(_.r)
          ],
          p: (C = f, g = _.p ?? f, C & -939524097 | g)
        };
      }) : Gr(b, l, c, f, u)
    ) : (
      // propagate className as is
      [
        {
          c: Ut(l),
          p: 0,
          o: 0,
          r: []
        }
      ]
    );
  }(a, t, r, n, i)) o.splice(Kr(o, s), 0, s);
  return o;
}
function Gt(e, t, r, n, i, o, a, s) {
  return Yr((s ? r.flatMap((l) => Je([
    l
  ], n, i, o, a)) : Je(r, n, i, o, a)).map((l) => (
    // do not move defaults
    // move only rules with a name unless they are in the base layer
    l.p & M.o && (l.n || t == M.b) ? {
      ...l,
      p: l.p & -939524097 | t,
      o: 0
    } : l
  )), e);
}
function mi(e, t, r, n) {
  var i;
  return i = (o, a) => {
    let { n: s, p: l, r: c, i: f } = Kt(o, a, t);
    return r && Gt(s, t, r, a, l, c, f, n);
  }, qr.set(e, i), e;
}
function ut(e, t, r) {
  if (e[e.length - 1] != "(") {
    let n = [], i = !1, o = !1, a = "";
    for (let s of e) if (!(s == "(" || /[~@]$/.test(s))) {
      if (s[0] == "!" && (s = s.slice(1), i = !i), s.endsWith(":")) {
        n[s == "dark:" ? "unshift" : "push"](s.slice(0, -1));
        continue;
      }
      s[0] == "-" && (s = s.slice(1), o = !o), s.endsWith("-") && (s = s.slice(0, -1)), s && s != "&" && (a += (a && "-") + s);
    }
    a && (o && (a = "-" + a), t[0].push({
      n: a,
      v: n.filter(bi),
      i
    }));
  }
}
function bi(e, t, r) {
  return r.indexOf(e) == t;
}
let vr = /* @__PURE__ */ new Map();
function Ge(e) {
  let t = vr.get(e);
  if (!t) {
    let r = [], n = [
      []
    ], i = 0, o = 0, a = null, s = 0, l = (c, f = 0) => {
      i != s && (r.push(e.slice(i, s + f)), c && ut(r, n)), i = s + 1;
    };
    for (; s < e.length; s++) {
      let c = e[s];
      if (o) e[s - 1] != "\\" && (o += +(c == "[") || -(c == "]"));
      else if (c == "[")
        o += 1;
      else if (a)
        e[s - 1] != "\\" && a.test(e.slice(s)) && (a = null, i = s + RegExp.lastMatch.length);
      else if (c == "/" && e[s - 1] != "\\" && (e[s + 1] == "*" || e[s + 1] == "/"))
        a = e[s + 1] == "*" ? /^\*\// : /^[\r\n]/;
      else if (c == "(")
        l(), r.push(c);
      else if (c == ":") e[s + 1] != ":" && l(!1, 1);
      else if (/[\s,)]/.test(c)) {
        l(!0);
        let f = r.lastIndexOf("(");
        if (c == ")") {
          let u = r[f - 1];
          if (/[~@]$/.test(u)) {
            let h = n.shift();
            r.length = f, ut([
              ...r,
              "#"
            ], n);
            let { v: b } = n[0].pop();
            for (let _ of h)
              _.v.splice(+(_.v[0] == "dark") - +(b[0] == "dark"), b.length);
            ut([
              ...r,
              mi(
                // named nested
                u.length > 1 ? u.slice(0, -1) + Ke(JSON.stringify([
                  u,
                  h
                ])) : u + "(" + hi(h) + ")",
                M.a,
                h,
                /@$/.test(u)
              )
            ], n);
          }
          f = r.lastIndexOf("(", f - 1);
        }
        r.length = f + 1;
      } else /[~@]/.test(c) && e[s + 1] == "(" && // start nested block
      // ~(...) or button~(...)
      // @(...) or button@(...)
      n.unshift([]);
    }
    l(!0), vr.set(e, t = n[0]);
  }
  return t;
}
function d(e, t, r) {
  return [
    e,
    vt(t, r)
  ];
}
function vt(e, t) {
  return typeof e == "function" ? e : typeof e == "string" && /^[\w-]+$/.test(e) ? (
    // a CSS property alias
    (r, n) => ({
      [e]: t ? t(r, n) : St(r, 1)
    })
  ) : (r) => (
    // CSSObject, shortcut or apply
    e || {
      [r[1]]: St(r, 2)
    }
  );
}
function St(e, t, r = e.slice(t).find(Boolean) || e.$$ || e.input) {
  return e.input[0] == "-" ? `calc(${r} * -1)` : r;
}
function p(e, t, r, n) {
  return [
    e,
    yi(t, r, n)
  ];
}
function yi(e, t, r) {
  let n = typeof t == "string" ? (i, o) => ({
    [t]: r ? r(i, o) : i._
  }) : t || (({ 1: i, _: o }, a, s) => ({
    [i || s]: o
  }));
  return (i, o) => {
    let a = Xr(e || i[1]), s = o.theme(a, i.$$) ?? Q(i.$$, a, o);
    if (s != null) return i._ = St(i, 0, s), n(i, o, a);
  };
}
function $(e, t = {}, r) {
  return [
    e,
    xi(t, r)
  ];
}
function xi(e = {}, t) {
  return (r, n) => {
    let { section: i = Xr(r[0]).replace("-", "") + "Color" } = e, [o, a] = wi(r.$$);
    if (!o) return;
    let s = n.theme(i, o) || Q(o, i, n);
    if (!s || typeof s == "object") return;
    let {
      // text- -> --tw-text-opacity
      // ring-offset(?:-|$) -> --tw-ring-offset-opacity
      // TODO move this default into preset-tailwind?
      opacityVariable: l = `--tw-${r[0].replace(/-$/, "")}-opacity`,
      opacitySection: c = i.replace("Color", "Opacity"),
      property: f = i,
      selector: u
    } = e, h = n.theme(c, a || "DEFAULT") || a && Q(a, c, n), b = t || (({ _: C }) => {
      let g = Ue(f, C);
      return u ? {
        [u]: g
      } : g;
    });
    r._ = {
      value: Z(s, {
        opacityVariable: l || void 0,
        opacityValue: h || void 0
      }),
      color: (C) => Z(s, C),
      opacityVariable: l || void 0,
      opacityValue: h || void 0
    };
    let _ = b(r, n);
    if (!r.dark) {
      let C = n.d(i, o, s);
      C && C !== s && (r._ = {
        value: Z(C, {
          opacityVariable: l || void 0,
          opacityValue: h || "1"
        }),
        color: (g) => Z(C, g),
        opacityVariable: l || void 0,
        opacityValue: h || void 0
      }, _ = {
        "&": _,
        [n.v("dark")]: b(r, n)
      });
    }
    return _;
  };
}
function wi(e) {
  return (e.match(/^(\[[^\]]+]|[^/]+?)(?:\/(.+))?$/) || []).slice(1);
}
function Ue(e, t) {
  let r = {};
  return typeof t == "string" ? r[e] = t : (t.opacityVariable && t.value.includes(t.opacityVariable) && (r[t.opacityVariable] = t.opacityValue || "1"), r[e] = t.value), r;
}
function Q(e, t, r) {
  if (e[0] == "[" && e.slice(-1) == "]") {
    if (e = je(Jt(e.slice(1, -1), r.theme)), !t) return e;
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
function Xr(e) {
  return e.replace(/-./g, (t) => t[1].toUpperCase());
}
function je(e) {
  return (
    // Keep raw strings if it starts with `url(`
    e.includes("url(") ? e.replace(/(.*?)(url\(.*?\))(.*?)/g, (t, r = "", n, i = "") => je(r) + n + je(i)) : e.replace(/(^|[^\\])_+/g, (t, r) => r + " ".repeat(t.length - r.length)).replace(/\\_/g, "_").replace(/(calc|min|max|clamp)\(.+\)/g, (t) => t.replace(/(-?\d*\.?\d(?!\b-.+[,)](?![^+\-/*])\D)(?:%|[a-z]+)?|\))([+\-/*])/g, "$1 $2 "))
  );
}
function Zr({ presets: e = [], ...t }) {
  let r = {
    darkMode: void 0,
    darkColor: void 0,
    preflight: t.preflight !== !1 && [],
    theme: {},
    variants: E(t.variants),
    rules: E(t.rules),
    ignorelist: E(t.ignorelist),
    hash: void 0,
    stringify: (n, i) => n + ":" + i,
    finalize: []
  };
  for (let n of E([
    ...e,
    {
      darkMode: t.darkMode,
      darkColor: t.darkColor,
      preflight: t.preflight !== !1 && E(t.preflight),
      theme: t.theme,
      hash: t.hash,
      stringify: t.stringify,
      finalize: t.finalize
    }
  ])) {
    let { preflight: i, darkMode: o = r.darkMode, darkColor: a = r.darkColor, theme: s, variants: l, rules: c, ignorelist: f, hash: u = r.hash, stringify: h = r.stringify, finalize: b } = typeof n == "function" ? n(r) : n;
    r = {
      // values defined by user or previous presets take precedence
      preflight: r.preflight !== !1 && i !== !1 && [
        ...r.preflight,
        ...E(i)
      ],
      darkMode: o,
      darkColor: a,
      theme: {
        ...r.theme,
        ...s,
        extend: {
          ...r.theme.extend,
          ...s == null ? void 0 : s.extend
        }
      },
      variants: [
        ...r.variants,
        ...E(l)
      ],
      rules: [
        ...r.rules,
        ...E(c)
      ],
      ignorelist: [
        ...r.ignorelist,
        ...E(f)
      ],
      hash: u,
      stringify: h,
      finalize: [
        ...r.finalize,
        ...E(b)
      ]
    };
  }
  return r;
}
function Sr(e, t, r, n, i, o) {
  for (let a of t) {
    let s = r.get(a);
    s || r.set(a, s = n(a));
    let l = s(e, i, o);
    if (l) return l;
  }
}
function _i(e) {
  var t;
  return Et(e[0], typeof (t = e[1]) == "function" ? t : () => t);
}
function vi(e) {
  var t, r;
  return Array.isArray(e) ? Et(e[0], vt(e[1], e[2])) : Et(e, vt(t, r));
}
function Et(e, t) {
  return Qr(e, (r, n, i, o) => {
    let a = n.exec(r);
    if (a) return (
      // MATCH.$_ = value
      a.$$ = r.slice(a[0].length), a.dark = o, t(a, i)
    );
  });
}
function Qr(e, t) {
  let r = E(e).map(Si);
  return (n, i, o) => {
    for (let a of r) {
      let s = t(n, a, i, o);
      if (s) return s;
    }
  };
}
function Si(e) {
  return typeof e == "string" ? RegExp("^" + e + (e.includes("$") || e.slice(-1) == "-" ? "" : "$")) : e;
}
function Ei(e, t) {
  let r = Zr(e), n = function({ theme: l, darkMode: c, darkColor: f = qt, variants: u, rules: h, hash: b, stringify: _, ignorelist: C, finalize: g }) {
    let w = /* @__PURE__ */ new Map(), m = /* @__PURE__ */ new Map(), y = /* @__PURE__ */ new Map(), S = /* @__PURE__ */ new Map(), q = Qr(C, (v, T) => T.test(v));
    u.push([
      "dark",
      Array.isArray(c) || c == "class" ? `${E(c)[1] || ".dark"} &` : typeof c == "string" && c != "media" ? c : (
        // a custom selector
        "@media (prefers-color-scheme:dark)"
      )
    ]);
    let P = typeof b == "function" ? (v) => b(v, Ke) : b ? Ke : yr;
    P !== yr && g.push((v) => {
      var T;
      return {
        ...v,
        n: v.n && P(v.n),
        d: (T = v.d) == null ? void 0 : T.replace(/--(tw(?:-[\w-]+)?)\b/g, (I, st) => "--" + P(st).replace("#", ""))
      };
    });
    let F = {
      theme: function({ extend: v = {}, ...T }) {
        let I = {}, st = {
          get colors() {
            return ve("colors");
          },
          theme: ve,
          // Stub implementation as negated values are automatically infered and do _not_ need to be in the theme
          negative() {
            return {};
          },
          breakpoints(R) {
            let z = {};
            for (let L in R) typeof R[L] == "string" && (z["screen-" + L] = R[L]);
            return z;
          }
        };
        return ve;
        function ve(R, z, L, Se) {
          if (R) {
            if ({ 1: R, 2: Se } = // eslint-disable-next-line no-sparse-arrays
            /^(\S+?)(?:\s*\/\s*([^/]+))?$/.exec(R) || [
              ,
              R
            ], /[.[]/.test(R)) {
              let B = [];
              R.replace(/\[([^\]]+)\]|([^.[]+)/g, (ie, Ae, pi = Ae) => B.push(pi)), R = B.shift(), L = z, z = B.join("-");
            }
            let K = I[R] || // two-step deref to allow extend section to reference base section
            Object.assign(Object.assign(
              // Make sure to not get into recursive calls
              I[R] = {},
              br(T, R)
            ), br(v, R));
            if (z == null) return K;
            z || (z = "DEFAULT");
            let ne = K[z] ?? z.split("-").reduce((B, ie) => B == null ? void 0 : B[ie], K) ?? L;
            return Se ? Z(ne, {
              opacityValue: Jt(Se, ve)
            }) : ne;
          }
          let Ee = {};
          for (let K of [
            ...Object.keys(T),
            ...Object.keys(v)
          ]) Ee[K] = ve(K);
          return Ee;
        }
        function br(R, z) {
          let L = R[z];
          return typeof L == "function" && (L = L(st)), L && /color|fill|stroke/i.test(z) ? function Se(Ee, K = []) {
            let ne = {};
            for (let B in Ee) {
              let ie = Ee[B], Ae = [
                ...K,
                B
              ];
              ne[Ae.join("-")] = ie, B == "DEFAULT" && (Ae = K, ne[K.join("-")] = ie), typeof ie == "object" && Object.assign(ne, Se(ie, Ae));
            }
            return ne;
          }(L) : L;
        }
      }(l),
      e: Ur,
      h: P,
      s(v, T) {
        return _(v, T, F);
      },
      d(v, T, I) {
        return f(v, T, F, I);
      },
      v(v) {
        return w.has(v) || w.set(v, Sr(v, u, m, _i, F) || "&:" + v), w.get(v);
      },
      r(v, T) {
        let I = JSON.stringify([
          v,
          T
        ]);
        return y.has(I) || y.set(I, !q(v, F) && Sr(v, h, S, vi, F, T)), y.get(I);
      },
      f(v) {
        return g.reduce((T, I) => I(T, F), v);
      }
    };
    return F;
  }(r), i = /* @__PURE__ */ new Map(), o = [], a = /* @__PURE__ */ new Set();
  t.resume((l) => i.set(l, l), (l, c) => {
    t.insert(l, o.length, c), o.push(c), a.add(l);
  });
  function s(l) {
    let c = n.f(l), f = _t(c);
    if (f && !a.has(f)) {
      a.add(f);
      let u = Kr(o, l);
      t.insert(f, u, l), o.splice(u, 0, l);
    }
    return c.n;
  }
  return Object.defineProperties(function(c) {
    if (!i.size) for (let u of E(r.preflight))
      typeof u == "function" && (u = u(n)), u && (typeof u == "string" ? Gt("", M.b, Ge(u), n, M.b, [], !1, !0) : Gr(u, {}, n, M.b)).forEach(s);
    c = "" + c;
    let f = i.get(c);
    if (!f) {
      let u = /* @__PURE__ */ new Set();
      for (let h of Je(Ge(c), n)) u.add(h.c).add(s(h));
      f = [
        ...u
      ].filter(Boolean).join(" "), i.set(c, f).set(f, f);
    }
    return f;
  }, Object.getOwnPropertyDescriptors({
    get target() {
      return t.target;
    },
    theme: n.theme,
    config: r,
    snapshot() {
      let l = t.snapshot(), c = new Set(a), f = new Map(i), u = [
        ...o
      ];
      return () => {
        l(), a = c, i = f, o = u;
      };
    },
    clear() {
      t.clear(), a = /* @__PURE__ */ new Set(), i = /* @__PURE__ */ new Map(), o = [];
    },
    destroy() {
      this.clear(), t.destroy();
    }
  }));
}
function Ai(e, t) {
  return e != t && "" + e.split(" ").sort() != "" + t.split(" ").sort();
}
function Ci(e) {
  let t = new MutationObserver(r);
  return {
    observe(i) {
      t.observe(i, {
        attributeFilter: [
          "class"
        ],
        subtree: !0,
        childList: !0
      }), n(i), r([
        {
          target: i,
          type: ""
        }
      ]);
    },
    disconnect() {
      t.disconnect();
    }
  };
  function r(i) {
    for (let { type: o, target: a } of i) if (o[0] == "a")
      n(a);
    else
      for (let s of a.querySelectorAll("[class]")) n(s);
    t.takeRecords();
  }
  function n(i) {
    var s;
    let o, a = (s = i.getAttribute) == null ? void 0 : s.call(i, "class");
    a && Ai(a, o = e(a)) && // Not using `target.className = ...` as that is read-only for SVGElements
    i.setAttribute("class", o);
  }
}
function ki(e) {
  let t = document.querySelector(e || 'style[data-twind=""]');
  return (!t || t.tagName != "STYLE") && (t = document.createElement("style"), document.head.prepend(t)), t.dataset.twind = "claimed", t;
}
function dt(e) {
  let t = e != null && e.cssRules ? e : (e && typeof e != "string" ? e : ki(e)).sheet;
  return {
    target: t,
    snapshot() {
      let r = Array.from(t.cssRules, (n) => n.cssText);
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
    insert(r, n) {
      try {
        t.insertRule(r, n);
      } catch {
        t.insertRule(":root{}", n);
      }
    },
    resume: qt
  };
}
function Oi(e, t = !0) {
  let r = function() {
    if (Ti) try {
      let l = dt(new CSSStyleSheet());
      return l.connect = (c) => {
        let f = pt(c);
        f.adoptedStyleSheets = [
          ...f.adoptedStyleSheets,
          l.target
        ];
      }, l.disconnect = qt, l;
    } catch {
    }
    let o = document.createElement("style");
    o.media = "not all", document.head.prepend(o);
    let a = [
      dt(o)
    ], s = /* @__PURE__ */ new WeakMap();
    return {
      get target() {
        return a[0].target;
      },
      snapshot() {
        let l = a.map((c) => c.snapshot());
        return () => l.forEach((c) => c());
      },
      clear() {
        a.forEach((l) => l.clear());
      },
      destroy() {
        a.forEach((l) => l.destroy());
      },
      insert(l, c, f) {
        a[0].insert(l, c, f);
        let u = this.target.cssRules[c];
        a.forEach((h, b) => b && h.target.insertRule(u.cssText, c));
      },
      resume(l, c) {
        return a[0].resume(l, c);
      },
      connect(l) {
        let c = document.createElement("style");
        pt(l).appendChild(c);
        let f = dt(c), { cssRules: u } = this.target;
        for (let h = 0; h < u.length; h++) f.target.insertRule(u[h].cssText, h);
        a.push(f), s.set(l, f);
      },
      disconnect(l) {
        let c = a.indexOf(s.get(l));
        c >= 0 && a.splice(c, 1);
      }
    };
  }(), n = Ei({
    ...e,
    // in production use short hashed class names
    hash: e.hash ?? t
  }, r), i = Ci(n);
  return function(a) {
    return class extends a {
      connectedCallback() {
        var l;
        (l = super.connectedCallback) == null || l.call(this), r.connect(this), i.observe(pt(this));
      }
      disconnectedCallback() {
        var l;
        r.disconnect(this), (l = super.disconnectedCallback) == null || l.call(this);
      }
      constructor(...l) {
        super(...l), this.tw = n;
      }
    };
  };
}
let Ti = typeof ShadowRoot < "u" && (typeof ShadyCSS > "u" || ShadyCSS.nativeShadow) && "adoptedStyleSheets" in Document.prototype && "replace" in CSSStyleSheet.prototype;
function pt(e) {
  return e.shadowRoot || e.attachShadow({
    mode: "open"
  });
}
var $i = /* @__PURE__ */ new Map([["align-self", "-ms-grid-row-align"], ["color-adjust", "-webkit-print-color-adjust"], ["column-gap", "grid-column-gap"], ["forced-color-adjust", "-ms-high-contrast-adjust"], ["gap", "grid-gap"], ["grid-template-columns", "-ms-grid-columns"], ["grid-template-rows", "-ms-grid-rows"], ["justify-self", "-ms-grid-column-align"], ["margin-inline-end", "-webkit-margin-end"], ["margin-inline-start", "-webkit-margin-start"], ["mask-border", "-webkit-mask-box-image"], ["mask-border-outset", "-webkit-mask-box-image-outset"], ["mask-border-slice", "-webkit-mask-box-image-slice"], ["mask-border-source", "-webkit-mask-box-image-source"], ["mask-border-repeat", "-webkit-mask-box-image-repeat"], ["mask-border-width", "-webkit-mask-box-image-width"], ["overflow-wrap", "word-wrap"], ["padding-inline-end", "-webkit-padding-end"], ["padding-inline-start", "-webkit-padding-start"], ["print-color-adjust", "color-adjust"], ["row-gap", "grid-row-gap"], ["scroll-margin-bottom", "scroll-snap-margin-bottom"], ["scroll-margin-left", "scroll-snap-margin-left"], ["scroll-margin-right", "scroll-snap-margin-right"], ["scroll-margin-top", "scroll-snap-margin-top"], ["scroll-margin", "scroll-snap-margin"], ["text-combine-upright", "-ms-text-combine-horizontal"]]);
function Ri(e) {
  return $i.get(e);
}
function Mi(e) {
  var t = /^(?:(text-(?:decoration$|e|or|si)|back(?:ground-cl|d|f)|box-d|mask(?:$|-[ispro]|-cl)|pr|hyphena|flex-d)|(tab-|column(?!-s)|text-align-l)|(ap)|u|hy)/i.exec(e);
  return t ? t[1] ? 1 : t[2] ? 2 : t[3] ? 3 : 5 : 0;
}
function ji(e, t) {
  var r = /^(?:(pos)|(cli)|(background-i)|(flex(?:$|-b)|(?:max-|min-)?(?:block-s|inl|he|widt))|dis)/i.exec(e);
  return r ? r[1] ? /^sti/i.test(t) ? 1 : 0 : r[2] ? /^pat/i.test(t) ? 1 : 0 : r[3] ? /^image-/i.test(t) ? 1 : 0 : r[4] ? t[3] === "-" ? 2 : 0 : /^(?:inline-)?grid$/i.test(t) ? 4 : 0 : 0;
}
let Fi = [
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
function Ii() {
  return ({ stringify: e }) => ({
    stringify(t, r, n) {
      let i = "", o = Ri(t);
      o && (i += e(o, r, n) + ";");
      let a = Mi(t), s = ji(t, r);
      for (let l of Fi)
        a & l[1] && (i += e(l[0] + t, r, n) + ";"), s & l[1] && (i += e(t, l[0] + r, n) + ";");
      return i + e(t, r, n);
    }
  });
}
let At = {
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
    .../* @__PURE__ */ D(4, "rem", 4, 0.5, 0.5),
    // 0.5: '0.125rem',
    // 1: '0.25rem',
    // 1.5: '0.375rem',
    // 2: '0.5rem',
    // 2.5: '0.625rem',
    // 3: '0.75rem',
    // 3.5: '0.875rem',
    // 4: '1rem',
    .../* @__PURE__ */ D(12, "rem", 4, 5),
    // 5: '1.25rem',
    // 6: '1.5rem',
    // 7: '1.75rem',
    // 8: '2rem',
    // 9: '2.25rem',
    // 10: '2.5rem',
    // 11: '2.75rem',
    // 12: '3rem',
    14: "3.5rem",
    .../* @__PURE__ */ D(64, "rem", 4, 16, 4),
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
  backdropBlur: /* @__PURE__ */ x("blur"),
  backdropBrightness: /* @__PURE__ */ x("brightness"),
  backdropContrast: /* @__PURE__ */ x("contrast"),
  backdropGrayscale: /* @__PURE__ */ x("grayscale"),
  backdropHueRotate: /* @__PURE__ */ x("hueRotate"),
  backdropInvert: /* @__PURE__ */ x("invert"),
  backdropOpacity: /* @__PURE__ */ x("opacity"),
  backdropSaturate: /* @__PURE__ */ x("saturate"),
  backdropSepia: /* @__PURE__ */ x("sepia"),
  backgroundColor: /* @__PURE__ */ x("colors"),
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
  backgroundOpacity: /* @__PURE__ */ x("opacity"),
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
    .../* @__PURE__ */ D(200, "", 100, 0, 50),
    // 0: '0',
    // 50: '.5',
    // 150: '1.5',
    // 200: '2',
    .../* @__PURE__ */ D(110, "", 100, 90, 5),
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
  borderOpacity: /* @__PURE__ */ x("opacity"),
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
  borderSpacing: /* @__PURE__ */ x("spacing"),
  borderWidth: {
    DEFAULT: "1px",
    .../* @__PURE__ */ N(8, "px")
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
  boxShadowColor: x("colors"),
  // container: {},
  // cursor: {
  //   // Default values are handled by plugin
  // },
  caretColor: /* @__PURE__ */ x("colors"),
  accentColor: ({ theme: e }) => ({
    auto: "auto",
    ...e("colors")
  }),
  contrast: {
    .../* @__PURE__ */ D(200, "", 100, 0, 50),
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
  divideColor: /* @__PURE__ */ x("borderColor"),
  divideOpacity: /* @__PURE__ */ x("borderOpacity"),
  divideWidth: /* @__PURE__ */ x("borderWidth"),
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
    ...Ce(2, 6),
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
    ...Ce(12, 12),
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
  gap: /* @__PURE__ */ x("spacing"),
  gradientColorStops: /* @__PURE__ */ x("colors"),
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
    ...Ce(2, 6),
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
    ...Ce(2, 4),
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
    .../* @__PURE__ */ D(10, "rem", 4, 3),
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
    .../* @__PURE__ */ D(100, "", 100, 0, 10),
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
  padding: /* @__PURE__ */ x("spacing"),
  placeholderColor: /* @__PURE__ */ x("colors"),
  placeholderOpacity: /* @__PURE__ */ x("opacity"),
  outlineColor: /* @__PURE__ */ x("colors"),
  outlineOffset: /* @__PURE__ */ N(8, "px"),
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',,
  outlineWidth: /* @__PURE__ */ N(8, "px"),
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',,
  ringColor: ({ theme: e }) => ({
    ...e("colors"),
    DEFAULT: "#3b82f6"
  }),
  ringOffsetColor: /* @__PURE__ */ x("colors"),
  ringOffsetWidth: /* @__PURE__ */ N(8, "px"),
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
    .../* @__PURE__ */ N(8, "px")
  },
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  rotate: {
    .../* @__PURE__ */ N(2, "deg"),
    // 0: '0deg',
    // 1: '1deg',
    // 2: '2deg',
    .../* @__PURE__ */ N(12, "deg", 3),
    // 3: '3deg',
    // 6: '6deg',
    // 12: '12deg',
    .../* @__PURE__ */ N(180, "deg", 45)
  },
  // 45: '45deg',
  // 90: '90deg',
  // 180: '180deg',
  saturate: /* @__PURE__ */ D(200, "", 100, 0, 50),
  // 0: '0',
  // 50: '.5',
  // 100: '1',
  // 150: '1.5',
  // 200: '2',
  scale: {
    .../* @__PURE__ */ D(150, "", 100, 0, 50),
    // 0: '0',
    // 50: '.5',
    // 150: '1.5',
    .../* @__PURE__ */ D(110, "", 100, 90, 5),
    // 90: '.9',
    // 95: '.95',
    // 100: '1',
    // 105: '1.05',
    // 110: '1.1',
    75: "0.75",
    125: "1.25"
  },
  scrollMargin: /* @__PURE__ */ x("spacing"),
  scrollPadding: /* @__PURE__ */ x("spacing"),
  sepia: {
    0: "0",
    DEFAULT: "100%"
  },
  skew: {
    .../* @__PURE__ */ N(2, "deg"),
    // 0: '0deg',
    // 1: '1deg',
    // 2: '2deg',
    .../* @__PURE__ */ N(12, "deg", 3)
  },
  // 3: '3deg',
  // 6: '6deg',
  // 12: '12deg',
  space: /* @__PURE__ */ x("spacing"),
  stroke: ({ theme: e }) => ({
    ...e("colors"),
    none: "none"
  }),
  strokeWidth: /* @__PURE__ */ D(2),
  // 0: '0',
  // 1: '1',
  // 2: '2',,
  textColor: /* @__PURE__ */ x("colors"),
  textDecorationColor: /* @__PURE__ */ x("colors"),
  textDecorationThickness: {
    "from-font": "from-font",
    auto: "auto",
    .../* @__PURE__ */ N(8, "px")
  },
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  textUnderlineOffset: {
    auto: "auto",
    .../* @__PURE__ */ N(8, "px")
  },
  // 0: '0px',
  // 1: '1px',
  // 2: '2px',
  // 4: '4px',
  // 8: '8px',
  textIndent: /* @__PURE__ */ x("spacing"),
  textOpacity: /* @__PURE__ */ x("opacity"),
  // transformOrigin: {
  //   // The following are already handled by the plugin:
  //   // center, right, left, bottom, top
  //   // 'bottom-10px-right-20px' -> bottom 10px right 20px
  // },
  transitionDuration: ({ theme: e }) => ({
    ...e("durations"),
    DEFAULT: "150ms"
  }),
  transitionDelay: /* @__PURE__ */ x("durations"),
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
    ...Ce(2, 4),
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
    .../* @__PURE__ */ D(50, "", 1, 0, 10),
    // 0: '0',
    // 10: '10',
    // 20: '20',
    // 30: '30',
    // 40: '40',
    // 50: '50',
    auto: "auto"
  }
};
function Ce(e, t) {
  let r = {};
  do
    for (var n = 1; n < e; n++) r[`${n}/${e}`] = Number((n / e * 100).toFixed(6)) + "%";
  while (++e <= t);
  return r;
}
function N(e, t, r = 0) {
  let n = {};
  for (; r <= e; r = 2 * r || 1) n[r] = r + t;
  return n;
}
function D(e, t = "", r = 1, n = 0, i = 1, o = {}) {
  for (; n <= e; n += i) o[n] = n / r + t;
  return o;
}
function x(e) {
  return ({ theme: t }) => t(e);
}
let zi = {
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
    fontFamily: `theme(fontFamily.sans, ${At.fontFamily.sans})`,
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
    fontFamily: `theme(fontFamily.mono, ${At.fontFamily.mono})`,
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
}, Li = [
  /* arbitrary properties: [paint-order:markers] */
  d("\\[([-\\w]+):(.+)]", ({ 1: e, 2: t }, r) => ({
    "@layer overrides": {
      "&": {
        [e]: Q(`[${t}]`, "", r)
      }
    }
  })),
  /* Styling based on parent and peer state */
  d("(group|peer)([~/][^-[]+)?", ({ input: e }, { h: t }) => [
    {
      c: t(e)
    }
  ]),
  /* LAYOUT */
  p("aspect-", "aspectRatio"),
  d("container", (e, { theme: t }) => {
    let { screens: r = t("screens"), center: n, padding: i } = t("container"), o = {
      width: "100%",
      marginRight: n && "auto",
      marginLeft: n && "auto",
      ...a("xs")
    };
    for (let s in r) {
      let l = r[s];
      typeof l == "string" && (o[Ht(l)] = {
        "&": {
          maxWidth: l,
          ...a(s)
        }
      });
    }
    return o;
    function a(s) {
      let l = i && (typeof i == "string" ? i : i[s] || i.DEFAULT);
      if (l) return {
        paddingRight: l,
        paddingLeft: l
      };
    }
  }),
  // Content
  p("content-", "content", ({ _: e }) => ({
    "--tw-content": e,
    content: "var(--tw-content)"
  })),
  // Box Decoration Break
  d("(?:box-)?decoration-(slice|clone)", "boxDecorationBreak"),
  // Box Sizing
  d("box-(border|content)", "boxSizing", ({ 1: e }) => e + "-box"),
  // Display
  d("hidden", {
    display: "none"
  }),
  // Table Layout
  d("table-(auto|fixed)", "tableLayout"),
  d([
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
  d("isolate", "isolation"),
  // Object Fit
  d("object-(contain|cover|fill|none|scale-down)", "objectFit"),
  // Object Position
  p("object-", "objectPosition"),
  d("object-(top|bottom|center|(left|right)(-(top|bottom))?)", "objectPosition", De),
  // Overscroll Behavior
  d("overscroll(-[xy])?-(auto|contain|none)", ({ 1: e = "", 2: t }) => ({
    ["overscroll-behavior" + e]: t
  })),
  // Position
  d("(static|fixed|absolute|relative|sticky)", "position"),
  // Top / Right / Bottom / Left
  p("-?inset(-[xy])?(?:$|-)", "inset", ({ 1: e, _: t }) => ({
    top: e != "-x" && t,
    right: e != "-y" && t,
    bottom: e != "-x" && t,
    left: e != "-y" && t
  })),
  p("-?(top|bottom|left|right)(?:$|-)", "inset"),
  // Visibility
  d("(visible|collapse)", "visibility"),
  d("invisible", {
    visibility: "hidden"
  }),
  // Z-Index
  p("-?z-", "zIndex"),
  /* FLEXBOX */
  // Flex Direction
  d("flex-((row|col)(-reverse)?)", "flexDirection", Er),
  d("flex-(wrap|wrap-reverse|nowrap)", "flexWrap"),
  p("(flex-(?:grow|shrink))(?:$|-)"),
  /*, 'flex-grow' | flex-shrink */
  p("(flex)-"),
  /*, 'flex' */
  p("grow(?:$|-)", "flexGrow"),
  p("shrink(?:$|-)", "flexShrink"),
  p("basis-", "flexBasis"),
  p("-?(order)-"),
  /*, 'order' */
  "-?(order)-(\\d+)",
  /* GRID */
  // Grid Template Columns
  p("grid-cols-", "gridTemplateColumns"),
  d("grid-cols-(\\d+)", "gridTemplateColumns", Or),
  // Grid Column Start / End
  p("col-", "gridColumn"),
  d("col-(span)-(\\d+)", "gridColumn", kr),
  p("col-start-", "gridColumnStart"),
  d("col-start-(auto|\\d+)", "gridColumnStart"),
  p("col-end-", "gridColumnEnd"),
  d("col-end-(auto|\\d+)", "gridColumnEnd"),
  // Grid Template Rows
  p("grid-rows-", "gridTemplateRows"),
  d("grid-rows-(\\d+)", "gridTemplateRows", Or),
  // Grid Row Start / End
  p("row-", "gridRow"),
  d("row-(span)-(\\d+)", "gridRow", kr),
  p("row-start-", "gridRowStart"),
  d("row-start-(auto|\\d+)", "gridRowStart"),
  p("row-end-", "gridRowEnd"),
  d("row-end-(auto|\\d+)", "gridRowEnd"),
  // Grid Auto Flow
  d("grid-flow-((row|col)(-dense)?)", "gridAutoFlow", (e) => De(Er(e))),
  d("grid-flow-(dense)", "gridAutoFlow"),
  // Grid Auto Columns
  p("auto-cols-", "gridAutoColumns"),
  // Grid Auto Rows
  p("auto-rows-", "gridAutoRows"),
  // Gap
  p("gap-x(?:$|-)", "gap", "columnGap"),
  p("gap-y(?:$|-)", "gap", "rowGap"),
  p("gap(?:$|-)", "gap"),
  /* BOX ALIGNMENT */
  // Justify Items
  // Justify Self
  "(justify-(?:items|self))-",
  // Justify Content
  d("justify-", "justifyContent", Ar),
  // Align Content
  // Align Items
  // Align Self
  d("(content|items|self)-", (e) => ({
    ["align-" + e[1]]: Ar(e)
  })),
  // Place Content
  // Place Items
  // Place Self
  d("(place-(content|items|self))-", ({ 1: e, $$: t }) => ({
    [e]: ("wun".includes(t[3]) ? "space-" : "") + t
  })),
  /* SPACING */
  // Padding
  p("p([xytrbl])?(?:$|-)", "padding", he("padding")),
  // Margin
  p("-?m([xytrbl])?(?:$|-)", "margin", he("margin")),
  // Space Between
  p("-?space-(x|y)(?:$|-)", "space", ({ 1: e, _: t }) => ({
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
  d("space-(x|y)-reverse", ({ 1: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      [`--tw-space-${e}-reverse`]: "1"
    }
  })),
  /* SIZING */
  // Width
  p("w-", "width"),
  // Min-Width
  p("min-w-", "minWidth"),
  // Max-Width
  p("max-w-", "maxWidth"),
  // Height
  p("h-", "height"),
  // Min-Height
  p("min-h-", "minHeight"),
  // Max-Height
  p("max-h-", "maxHeight"),
  /* TYPOGRAPHY */
  // Font Weight
  p("font-", "fontWeight"),
  // Font Family
  p("font-", "fontFamily", ({ _: e }) => typeof (e = E(e))[1] == "string" ? {
    fontFamily: V(e)
  } : {
    fontFamily: V(e[0]),
    ...e[1]
  }),
  // Font Smoothing
  d("antialiased", {
    WebkitFontSmoothing: "antialiased",
    MozOsxFontSmoothing: "grayscale"
  }),
  d("subpixel-antialiased", {
    WebkitFontSmoothing: "auto",
    MozOsxFontSmoothing: "auto"
  }),
  // Font Style
  d("italic", "fontStyle"),
  d("not-italic", {
    fontStyle: "normal"
  }),
  // Font Variant Numeric
  d("(ordinal|slashed-zero|(normal|lining|oldstyle|proportional|tabular)-nums|(diagonal|stacked)-fractions)", ({ 1: e, 2: t = "", 3: r }) => (
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
      ...X({
        "--tw-ordinal": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-slashed-zero": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-numeric-figure": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-numeric-spacing": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-numeric-fraction": "var(--tw-empty,/*!*/ /*!*/)"
      })
    }
  )),
  // Letter Spacing
  p("tracking-", "letterSpacing"),
  // Line Height
  p("leading-", "lineHeight"),
  // List Style Position
  d("list-(inside|outside)", "listStylePosition"),
  // List Style Type
  p("list-", "listStyleType"),
  d("list-", "listStyleType"),
  // Placeholder Opacity
  p("placeholder-opacity-", "placeholderOpacity", ({ _: e }) => ({
    "&::placeholder": {
      "--tw-placeholder-opacity": e
    }
  })),
  // Placeholder Color
  $("placeholder-", {
    property: "color",
    selector: "&::placeholder"
  }),
  // Text Alignment
  d("text-(left|center|right|justify|start|end)", "textAlign"),
  d("text-(ellipsis|clip)", "textOverflow"),
  // Text Opacity
  p("text-opacity-", "textOpacity", "--tw-text-opacity"),
  // Text Color
  $("text-", {
    property: "color"
  }),
  // Font Size
  p("text-", "fontSize", ({ _: e }) => typeof e == "string" ? {
    fontSize: e
  } : {
    fontSize: e[0],
    ...typeof e[1] == "string" ? {
      lineHeight: e[1]
    } : e[1]
  }),
  // Text Indent
  p("indent-", "textIndent"),
  // Text Decoration
  d("(overline|underline|line-through)", "textDecorationLine"),
  d("no-underline", {
    textDecorationLine: "none"
  }),
  // Text Underline offset
  p("underline-offset-", "textUnderlineOffset"),
  // Text Decoration Color
  $("decoration-", {
    section: "textDecorationColor",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Text Decoration Thickness
  p("decoration-", "textDecorationThickness"),
  // Text Decoration Style
  d("decoration-", "textDecorationStyle"),
  // Text Transform
  d("(uppercase|lowercase|capitalize)", "textTransform"),
  d("normal-case", {
    textTransform: "none"
  }),
  // Text Overflow
  d("truncate", {
    overflow: "hidden",
    whiteSpace: "nowrap",
    textOverflow: "ellipsis"
  }),
  // Vertical Alignment
  d("align-", "verticalAlign"),
  // Whitespace
  d("whitespace-", "whiteSpace"),
  // Word Break
  d("break-normal", {
    wordBreak: "normal",
    overflowWrap: "normal"
  }),
  d("break-words", {
    overflowWrap: "break-word"
  }),
  d("break-all", {
    wordBreak: "break-all"
  }),
  d("break-keep", {
    wordBreak: "keep-all"
  }),
  // Caret Color
  $("caret-", {
    // section: 'caretColor',
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Accent Color
  $("accent-", {
    // section: 'accentColor',
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Gradient Color Stops
  d("bg-gradient-to-([trbl]|[tb][rl])", "backgroundImage", ({ 1: e }) => `linear-gradient(to ${ae(e, " ")},var(--tw-gradient-stops))`),
  $("from-", {
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
  $("via-", {
    section: "gradientColorStops",
    opacityVariable: !1,
    opacitySection: "opacity"
  }, ({ _: e }) => ({
    "--tw-gradient-to": e.color({
      opacityValue: "0"
    }),
    "--tw-gradient-stops": `var(--tw-gradient-from),${e.value},var(--tw-gradient-to)`
  })),
  $("to-", {
    section: "gradientColorStops",
    property: "--tw-gradient-to",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  /* BACKGROUNDS */
  // Background Attachment
  d("bg-(fixed|local|scroll)", "backgroundAttachment"),
  // Background Origin
  d("bg-origin-(border|padding|content)", "backgroundOrigin", ({ 1: e }) => e + "-box"),
  // Background Repeat
  d([
    "bg-(no-repeat|repeat(-[xy])?)",
    "bg-repeat-(round|space)"
  ], "backgroundRepeat"),
  // Background Blend Mode
  d("bg-blend-", "backgroundBlendMode"),
  // Background Clip
  d("bg-clip-(border|padding|content|text)", "backgroundClip", ({ 1: e }) => e + (e == "text" ? "" : "-box")),
  // Background Opacity
  p("bg-opacity-", "backgroundOpacity", "--tw-bg-opacity"),
  // Background Color
  // bg-${backgroundColor}/${backgroundOpacity}
  $("bg-", {
    section: "backgroundColor"
  }),
  // Background Image
  // supported arbitrary types are: length, color, angle, list
  p("bg-", "backgroundImage"),
  // Background Position
  p("bg-", "backgroundPosition"),
  d("bg-(top|bottom|center|(left|right)(-(top|bottom))?)", "backgroundPosition", De),
  // Background Size
  p("bg-", "backgroundSize"),
  /* BORDERS */
  // Border Radius
  p("rounded(?:$|-)", "borderRadius"),
  p("rounded-([trbl]|[tb][rl])(?:$|-)", "borderRadius", ({ 1: e, _: t }) => {
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
      [`border-${ae(r[0])}-radius`]: t,
      [`border-${ae(r[1])}-radius`]: t
    };
  }),
  // Border Collapse
  d("border-(collapse|separate)", "borderCollapse"),
  // Border Opacity
  p("border-opacity(?:$|-)", "borderOpacity", "--tw-border-opacity"),
  // Border Style
  d("border-(solid|dashed|dotted|double|none)", "borderStyle"),
  // Border Spacing
  p("border-spacing(-[xy])?(?:$|-)", "borderSpacing", ({ 1: e, _: t }) => ({
    ...X({
      "--tw-border-spacing-x": "0",
      "--tw-border-spacing-y": "0"
    }),
    ["--tw-border-spacing" + (e || "-x")]: t,
    ["--tw-border-spacing" + (e || "-y")]: t,
    "border-spacing": "var(--tw-border-spacing-x) var(--tw-border-spacing-y)"
  })),
  // Border Color
  $("border-([xytrbl])-", {
    section: "borderColor"
  }, he("border", "Color")),
  $("border-"),
  // Border Width
  p("border-([xytrbl])(?:$|-)", "borderWidth", he("border", "Width")),
  p("border(?:$|-)", "borderWidth"),
  // Divide Opacity
  p("divide-opacity(?:$|-)", "divideOpacity", ({ _: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      "--tw-divide-opacity": e
    }
  })),
  // Divide Style
  d("divide-(solid|dashed|dotted|double|none)", ({ 1: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      borderStyle: e
    }
  })),
  // Divide Width
  d("divide-([xy]-reverse)", ({ 1: e }) => ({
    "&>:not([hidden])~:not([hidden])": {
      ["--tw-divide-" + e]: "1"
    }
  })),
  p("divide-([xy])(?:$|-)", "divideWidth", ({ 1: e, _: t }) => {
    let r = {
      x: "lr",
      y: "tb"
    }[e];
    return {
      "&>:not([hidden])~:not([hidden])": {
        [`--tw-divide-${e}-reverse`]: "0",
        [`border-${ae(r[0])}Width`]: `calc(${t} * calc(1 - var(--tw-divide-${e}-reverse)))`,
        [`border-${ae(r[1])}Width`]: `calc(${t} * var(--tw-divide-${e}-reverse))`
      }
    };
  }),
  // Divide Color
  $("divide-", {
    // section: $0.replace('-', 'Color') -> 'divideColor'
    property: "borderColor",
    // opacityVariable: '--tw-border-opacity',
    // opacitySection: section.replace('Color', 'Opacity') -> 'divideOpacity'
    selector: "&>:not([hidden])~:not([hidden])"
  }),
  // Ring Offset Opacity
  p("ring-opacity(?:$|-)", "ringOpacity", "--tw-ring-opacity"),
  // Ring Offset Color
  $("ring-offset-", {
    // section: 'ringOffsetColor',
    property: "--tw-ring-offset-color",
    opacityVariable: !1
  }),
  // opacitySection: section.replace('Color', 'Opacity') -> 'ringOffsetOpacity'
  // Ring Offset Width
  p("ring-offset(?:$|-)", "ringOffsetWidth", "--tw-ring-offset-width"),
  // Ring Inset
  d("ring-inset", {
    "--tw-ring-inset": "inset"
  }),
  // Ring Color
  $("ring-", {
    // section: 'ringColor',
    property: "--tw-ring-color"
  }),
  // opacityVariable: '--tw-ring-opacity',
  // opacitySection: section.replace('Color', 'Opacity') -> 'ringOpacity'
  // Ring Width
  p("ring(?:$|-)", "ringWidth", ({ _: e }, { theme: t }) => ({
    ...X({
      "--tw-ring-offset-shadow": "0 0 #0000",
      "--tw-ring-shadow": "0 0 #0000",
      "--tw-shadow": "0 0 #0000",
      "--tw-shadow-colored": "0 0 #0000",
      // Within own declaration to have the defaults above to be merged with defaults from shadow
      "&": {
        "--tw-ring-inset": "var(--tw-empty,/*!*/ /*!*/)",
        "--tw-ring-offset-width": t("ringOffsetWidth", "", "0px"),
        "--tw-ring-offset-color": Z(t("ringOffsetColor", "", "#fff")),
        "--tw-ring-color": Z(t("ringColor", "", "#93c5fd"), {
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
  $("shadow-", {
    section: "boxShadowColor",
    opacityVariable: !1,
    opacitySection: "opacity"
  }, ({ _: e }) => ({
    "--tw-shadow-color": e.value,
    "--tw-shadow": "var(--tw-shadow-colored)"
  })),
  // Box Shadow
  p("shadow(?:$|-)", "boxShadow", ({ _: e }) => ({
    ...X({
      "--tw-ring-offset-shadow": "0 0 #0000",
      "--tw-ring-shadow": "0 0 #0000",
      "--tw-shadow": "0 0 #0000",
      "--tw-shadow-colored": "0 0 #0000"
    }),
    "--tw-shadow": V(e),
    // replace all colors with reference to --tw-shadow-colored
    // this matches colors after non-comma char (keyword, offset) before comma or the end
    "--tw-shadow-colored": V(e).replace(/([^,]\s+)(?:#[a-f\d]+|(?:(?:hsl|rgb)a?|hwb|lab|lch|color|var)\(.+?\)|[a-z]+)(,|$)/g, "$1var(--tw-shadow-color)$2"),
    boxShadow: "var(--tw-ring-offset-shadow),var(--tw-ring-shadow),var(--tw-shadow)"
  })),
  // Opacity
  p("(opacity)-"),
  /*, 'opacity' */
  // Mix Blend Mode
  d("mix-blend-", "mixBlendMode"),
  /* FILTERS */
  ...Cr(),
  ...Cr("backdrop-"),
  /* TRANSITIONS AND ANIMATION */
  // Transition Property
  p("transition(?:$|-)", "transitionProperty", (e, { theme: t }) => ({
    transitionProperty: V(e),
    transitionTimingFunction: e._ == "none" ? void 0 : V(t("transitionTimingFunction", "")),
    transitionDuration: e._ == "none" ? void 0 : V(t("transitionDuration", ""))
  })),
  // Transition Duration
  p("duration(?:$|-)", "transitionDuration", "transitionDuration", V),
  // Transition Timing Function
  p("ease(?:$|-)", "transitionTimingFunction", "transitionTimingFunction", V),
  // Transition Delay
  p("delay(?:$|-)", "transitionDelay", "transitionDelay", V),
  p("animate(?:$|-)", "animation", (e, { theme: t, h: r, e: n }) => {
    let i = V(e), o = i.split(" "), a = t("keyframes", o[0]);
    return a ? {
      ["@keyframes " + (o[0] = n(r(o[0])))]: a,
      animation: o.join(" ")
    } : {
      animation: i
    };
  }),
  /* TRANSFORMS */
  // Transform
  "(transform)-(none)",
  d("transform", Ct),
  d("transform-(cpu|gpu)", ({ 1: e }) => ({
    "--tw-transform": en(e == "gpu")
  })),
  // Scale
  p("scale(-[xy])?-", "scale", ({ 1: e, _: t }) => ({
    ["--tw-scale" + (e || "-x")]: t,
    ["--tw-scale" + (e || "-y")]: t,
    ...Ct()
  })),
  // Rotate
  p("-?(rotate)-", "rotate", ht),
  // Translate
  p("-?(translate-[xy])-", "translate", ht),
  // Skew
  p("-?(skew-[xy])-", "skew", ht),
  // Transform Origin
  d("origin-(center|((top|bottom)(-(left|right))?)|left|right)", "transformOrigin", De),
  /* INTERACTIVITY */
  // Appearance
  "(appearance)-",
  // Columns
  p("(columns)-"),
  /*, 'columns' */
  "(columns)-(\\d+)",
  // Break Before, After and Inside
  "(break-(?:before|after|inside))-",
  // Cursor
  p("(cursor)-"),
  /*, 'cursor' */
  "(cursor)-",
  // Scroll Snap Type
  d("snap-(none)", "scroll-snap-type"),
  d("snap-(x|y|both)", ({ 1: e }) => ({
    ...X({
      "--tw-scroll-snap-strictness": "proximity"
    }),
    "scroll-snap-type": e + " var(--tw-scroll-snap-strictness)"
  })),
  d("snap-(mandatory|proximity)", "--tw-scroll-snap-strictness"),
  // Scroll Snap Align
  d("snap-(?:(start|end|center)|align-(none))", "scroll-snap-align"),
  // Scroll Snap Stop
  d("snap-(normal|always)", "scroll-snap-stop"),
  d("scroll-(auto|smooth)", "scroll-behavior"),
  // Scroll Margin
  // Padding
  p("scroll-p([xytrbl])?(?:$|-)", "padding", he("scroll-padding")),
  // Margin
  p("-?scroll-m([xytrbl])?(?:$|-)", "scroll-margin", he("scroll-margin")),
  // Touch Action
  d("touch-(auto|none|manipulation)", "touch-action"),
  d("touch-(pinch-zoom|pan-(?:(x|left|right)|(y|up|down)))", ({ 1: e, 2: t, 3: r }) => ({
    ...X({
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
  d("outline-none", {
    outline: "2px solid transparent",
    "outline-offset": "2px"
  }),
  d("outline", {
    outlineStyle: "solid"
  }),
  d("outline-(dashed|dotted|double)", "outlineStyle"),
  // Outline Offset
  p("-?(outline-offset)-"),
  /*, 'outlineOffset'*/
  // Outline Color
  $("outline-", {
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Outline Width
  p("outline-", "outlineWidth"),
  // Pointer Events
  "(pointer-events)-",
  // Will Change
  p("(will-change)-"),
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
  d("select-(none|text|all|auto)", "userSelect"),
  /* SVG */
  // Fill, Stroke
  $("fill-", {
    section: "fill",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  $("stroke-", {
    section: "stroke",
    opacityVariable: !1,
    opacitySection: "opacity"
  }),
  // Stroke Width
  p("stroke-", "strokeWidth"),
  /* ACCESSIBILITY */
  // Screen Readers
  d("sr-only", {
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
  d("not-sr-only", {
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
function De(e) {
  return (typeof e == "string" ? e : e[1]).replace(/-/g, " ").trim();
}
function Er(e) {
  return (typeof e == "string" ? e : e[1]).replace("col", "column");
}
function ae(e, t = "-") {
  let r = [];
  for (let n of e) r.push({
    t: "top",
    r: "right",
    b: "bottom",
    l: "left"
  }[n]);
  return r.join(t);
}
function V(e) {
  return e && "" + (e._ || e);
}
function Ar({ $$: e }) {
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
function he(e, t = "") {
  return ({ 1: r, _: n }) => {
    let i = {
      x: "lr",
      y: "tb"
    }[r] || r + r;
    return i ? {
      ...Ue(e + "-" + ae(i[0]) + t, n),
      ...Ue(e + "-" + ae(i[1]) + t, n)
    } : Ue(e + t, n);
  };
}
function Cr(e = "") {
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
  for (let n of t) r[`--tw-${e}${n}`] = "var(--tw-empty,/*!*/ /*!*/)";
  return r = {
    // move defaults
    ...X(r),
    // add default filter which allows standalone usage
    [`${e}filter`]: t.map((n) => `var(--tw-${e}${n})`).join(" ")
  }, [
    `(${e}filter)-(none)`,
    d(`${e}filter`, r),
    ...t.map((n) => p(
      // hue-rotate can be negated
      `${n[0] == "h" ? "-?" : ""}(${e}${n})(?:$|-)`,
      n,
      ({ 1: i, _: o }) => ({
        [`--tw-${i}`]: E(o).map((a) => `${n}(${a})`).join(" "),
        ...r
      })
    ))
  ];
}
function ht({ 1: e, _: t }) {
  return {
    ["--tw-" + e]: t,
    ...Ct()
  };
}
function Ct() {
  return {
    ...X({
      "--tw-translate-x": "0",
      "--tw-translate-y": "0",
      "--tw-rotate": "0",
      "--tw-skew-x": "0",
      "--tw-skew-y": "0",
      "--tw-scale-x": "1",
      "--tw-scale-y": "1",
      "--tw-transform": en()
    }),
    transform: "var(--tw-transform)"
  };
}
function en(e) {
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
function kr({ 1: e, 2: t }) {
  return `${e} ${t} / ${e} ${t}`;
}
function Or({ 1: e }) {
  return `repeat(${e},minmax(0,1fr))`;
}
function X(e) {
  return {
    "@layer defaults": {
      "*,::before,::after": e,
      "::backdrop": e
    }
  };
}
let Di = [
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
    Q(t, "", r) || // default handling
    `${t}="true"`}]`
  ],
  /* Styling based on parent and peer state */
  // Groups classes like: group-focus and group-hover
  // these need to add a marker selector with the pseudo class
  // => '.group:focus .group-focus:selector'
  [
    "((group|peer)(~[^-[]+)?)(-\\[(.+)]|[-[].+?)(\\/.+)?",
    ({ 2: e, 3: t = "", 4: r, 5: n = "", 6: i = t }, { e: o, h: a, v: s }) => {
      let l = je(n) || (r[0] == "[" ? r : s(r.slice(1)));
      return `${(l.includes("&") ? l : "&" + l).replace(/&/g, `:merge(.${o(a(e + i))})`)}${e[0] == "p" ? "~" : " "}&`;
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
      if (e && (e = t.theme("supports", e) || Q(e, "", t)), e) return e.includes(":") || (e += ":var(--tw)"), /^\w*\s*\(/.test(e) || (e = `(${e})`), // Chrome has a bug where `(condtion1)or(condition2)` is not valid
      // But `(condition1) or (condition2)` is supported.
      `@supports ${e.replace(/\b(and|or|not)\b/g, " $1 ").trim()}`;
    }
  ],
  [
    "max-",
    ({ $$: e }, t) => {
      if (e && (e = t.theme("screens", e) || Q(e, "", t)), typeof e == "string") return `@media not all and (min-width:${e})`;
    }
  ],
  [
    "min-",
    ({ $$: e }, t) => (e && (e = Q(e, "", t)), e && `@media (min-width:${e})`)
  ],
  // Arbitrary variants
  [
    /^\[(.+)]$/,
    ({ 1: e }) => /[&@]/.test(e) && je(e).replace(/[}]+$/, "").split("{")
  ]
];
function Pi({ colors: e, disablePreflight: t } = {}) {
  return {
    // allow other preflight to run
    preflight: t ? void 0 : zi,
    theme: {
      ...At,
      colors: {
        inherit: "inherit",
        current: "currentColor",
        transparent: "transparent",
        black: "#000",
        white: "#fff",
        ...e
      }
    },
    variants: Di,
    rules: Li,
    finalize(r) {
      return (
        // automatically add `content: ''` to before and after so you don’t have to specify it unless you want a different value
        // ignore global, preflight, and auto added rules
        r.n && // only if there are declarations
        r.d && // and it has a ::before or ::after selector
        r.r.some((n) => /^&::(before|after)$/.test(n)) && // there is no content property yet
        !/(^|;)content:/.test(r.d) ? {
          ...r,
          d: "content:var(--tw-content);" + r.d
        } : r
      );
    }
  };
}
let Ni = {
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
}, Wi = {
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
}, Bi = {
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
}, Vi = {
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
}, Ui = {
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
}, Hi = {
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
}, qi = {
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
}, Ki = {
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
}, Ji = {
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
}, Gi = {
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
}, Yi = {
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
}, Xi = {
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
}, Zi = {
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
}, Qi = {
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
}, eo = {
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
}, to = {
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
}, ro = {
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
}, no = {
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
}, io = {
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
}, oo = {
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
}, ao = {
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
}, so = {
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
}, lo = {
  __proto__: null,
  slate: Ni,
  gray: Wi,
  zinc: Bi,
  neutral: Vi,
  stone: Ui,
  red: Hi,
  orange: qi,
  amber: Ki,
  yellow: Ji,
  lime: Gi,
  green: Yi,
  emerald: Xi,
  teal: Zi,
  cyan: Qi,
  sky: eo,
  blue: to,
  indigo: ro,
  violet: no,
  purple: io,
  fuchsia: oo,
  pink: ao,
  rose: so
};
function co({ disablePreflight: e } = {}) {
  return Pi({
    colors: lo,
    disablePreflight: e
  });
}
var kt = !1, Ot = !1, le = [], Tt = -1;
function fo(e) {
  uo(e);
}
function uo(e) {
  le.includes(e) || le.push(e), ho();
}
function po(e) {
  let t = le.indexOf(e);
  t !== -1 && t > Tt && le.splice(t, 1);
}
function ho() {
  !Ot && !kt && (kt = !0, queueMicrotask(go));
}
function go() {
  kt = !1, Ot = !0;
  for (let e = 0; e < le.length; e++)
    le[e](), Tt = e;
  le.length = 0, Tt = -1, Ot = !1;
}
var be, pe, ye, tn, $t = !0;
function mo(e) {
  $t = !1, e(), $t = !0;
}
function bo(e) {
  be = e.reactive, ye = e.release, pe = (t) => e.effect(t, { scheduler: (r) => {
    $t ? fo(r) : r();
  } }), tn = e.raw;
}
function Tr(e) {
  pe = e;
}
function yo(e) {
  let t = () => {
  };
  return [(n) => {
    let i = pe(n);
    return e._x_effects || (e._x_effects = /* @__PURE__ */ new Set(), e._x_runEffects = () => {
      e._x_effects.forEach((o) => o());
    }), e._x_effects.add(i), t = () => {
      i !== void 0 && (e._x_effects.delete(i), ye(i));
    }, i;
  }, () => {
    t();
  }];
}
function rn(e, t) {
  let r = !0, n, i = pe(() => {
    let o = e();
    JSON.stringify(o), r ? n = o : queueMicrotask(() => {
      t(o, n), n = o;
    }), r = !1;
  });
  return () => ye(i);
}
var nn = [], on = [], an = [];
function xo(e) {
  an.push(e);
}
function Yt(e, t) {
  typeof t == "function" ? (e._x_cleanups || (e._x_cleanups = []), e._x_cleanups.push(t)) : (t = e, on.push(t));
}
function sn(e) {
  nn.push(e);
}
function ln(e, t, r) {
  e._x_attributeCleanups || (e._x_attributeCleanups = {}), e._x_attributeCleanups[t] || (e._x_attributeCleanups[t] = []), e._x_attributeCleanups[t].push(r);
}
function cn(e, t) {
  e._x_attributeCleanups && Object.entries(e._x_attributeCleanups).forEach(([r, n]) => {
    (t === void 0 || t.includes(r)) && (n.forEach((i) => i()), delete e._x_attributeCleanups[r]);
  });
}
function wo(e) {
  var t, r;
  for ((t = e._x_effects) == null || t.forEach(po); (r = e._x_cleanups) != null && r.length; )
    e._x_cleanups.pop()();
}
var Xt = new MutationObserver(tr), Zt = !1;
function Qt() {
  Xt.observe(document, { subtree: !0, childList: !0, attributes: !0, attributeOldValue: !0 }), Zt = !0;
}
function fn() {
  _o(), Xt.disconnect(), Zt = !1;
}
var ke = [];
function _o() {
  let e = Xt.takeRecords();
  ke.push(() => e.length > 0 && tr(e));
  let t = ke.length;
  queueMicrotask(() => {
    if (ke.length === t)
      for (; ke.length > 0; )
        ke.shift()();
  });
}
function k(e) {
  if (!Zt)
    return e();
  fn();
  let t = e();
  return Qt(), t;
}
var er = !1, Ye = [];
function vo() {
  er = !0;
}
function So() {
  er = !1, tr(Ye), Ye = [];
}
function tr(e) {
  if (er) {
    Ye = Ye.concat(e);
    return;
  }
  let t = [], r = /* @__PURE__ */ new Set(), n = /* @__PURE__ */ new Map(), i = /* @__PURE__ */ new Map();
  for (let o = 0; o < e.length; o++)
    if (!e[o].target._x_ignoreMutationObserver && (e[o].type === "childList" && (e[o].removedNodes.forEach((a) => {
      a.nodeType === 1 && a._x_marker && r.add(a);
    }), e[o].addedNodes.forEach((a) => {
      if (a.nodeType === 1) {
        if (r.has(a)) {
          r.delete(a);
          return;
        }
        a._x_marker || t.push(a);
      }
    })), e[o].type === "attributes")) {
      let a = e[o].target, s = e[o].attributeName, l = e[o].oldValue, c = () => {
        n.has(a) || n.set(a, []), n.get(a).push({ name: s, value: a.getAttribute(s) });
      }, f = () => {
        i.has(a) || i.set(a, []), i.get(a).push(s);
      };
      a.hasAttribute(s) && l === null ? c() : a.hasAttribute(s) ? (f(), c()) : f();
    }
  i.forEach((o, a) => {
    cn(a, o);
  }), n.forEach((o, a) => {
    nn.forEach((s) => s(a, o));
  });
  for (let o of r)
    t.some((a) => a.contains(o)) || on.forEach((a) => a(o));
  for (let o of t)
    o.isConnected && an.forEach((a) => a(o));
  t = null, r = null, n = null, i = null;
}
function un(e) {
  return ze(ge(e));
}
function Ie(e, t, r) {
  return e._x_dataStack = [t, ...ge(r || e)], () => {
    e._x_dataStack = e._x_dataStack.filter((n) => n !== t);
  };
}
function ge(e) {
  return e._x_dataStack ? e._x_dataStack : typeof ShadowRoot == "function" && e instanceof ShadowRoot ? ge(e.host) : e.parentNode ? ge(e.parentNode) : [];
}
function ze(e) {
  return new Proxy({ objects: e }, Eo);
}
var Eo = {
  ownKeys({ objects: e }) {
    return Array.from(
      new Set(e.flatMap((t) => Object.keys(t)))
    );
  },
  has({ objects: e }, t) {
    return t == Symbol.unscopables ? !1 : e.some(
      (r) => Object.prototype.hasOwnProperty.call(r, t) || Reflect.has(r, t)
    );
  },
  get({ objects: e }, t, r) {
    return t == "toJSON" ? Ao : Reflect.get(
      e.find(
        (n) => Reflect.has(n, t)
      ) || {},
      t,
      r
    );
  },
  set({ objects: e }, t, r, n) {
    const i = e.find(
      (a) => Object.prototype.hasOwnProperty.call(a, t)
    ) || e[e.length - 1], o = Object.getOwnPropertyDescriptor(i, t);
    return o != null && o.set && (o != null && o.get) ? o.set.call(n, r) || !0 : Reflect.set(i, t, r);
  }
};
function Ao() {
  return Reflect.ownKeys(this).reduce((t, r) => (t[r] = Reflect.get(this, r), t), {});
}
function dn(e) {
  let t = (n) => typeof n == "object" && !Array.isArray(n) && n !== null, r = (n, i = "") => {
    Object.entries(Object.getOwnPropertyDescriptors(n)).forEach(([o, { value: a, enumerable: s }]) => {
      if (s === !1 || a === void 0 || typeof a == "object" && a !== null && a.__v_skip)
        return;
      let l = i === "" ? o : `${i}.${o}`;
      typeof a == "object" && a !== null && a._x_interceptor ? n[o] = a.initialize(e, l, o) : t(a) && a !== n && !(a instanceof Element) && r(a, l);
    });
  };
  return r(e);
}
function pn(e, t = () => {
}) {
  let r = {
    initialValue: void 0,
    _x_interceptor: !0,
    initialize(n, i, o) {
      return e(this.initialValue, () => Co(n, i), (a) => Rt(n, i, a), i, o);
    }
  };
  return t(r), (n) => {
    if (typeof n == "object" && n !== null && n._x_interceptor) {
      let i = r.initialize.bind(r);
      r.initialize = (o, a, s) => {
        let l = n.initialize(o, a, s);
        return r.initialValue = l, i(o, a, s);
      };
    } else
      r.initialValue = n;
    return r;
  };
}
function Co(e, t) {
  return t.split(".").reduce((r, n) => r[n], e);
}
function Rt(e, t, r) {
  if (typeof t == "string" && (t = t.split(".")), t.length === 1)
    e[t[0]] = r;
  else {
    if (t.length === 0)
      throw error;
    return e[t[0]] || (e[t[0]] = {}), Rt(e[t[0]], t.slice(1), r);
  }
}
var hn = {};
function H(e, t) {
  hn[e] = t;
}
function Mt(e, t) {
  let r = ko(t);
  return Object.entries(hn).forEach(([n, i]) => {
    Object.defineProperty(e, `$${n}`, {
      get() {
        return i(t, r);
      },
      enumerable: !1
    });
  }), e;
}
function ko(e) {
  let [t, r] = wn(e), n = { interceptor: pn, ...t };
  return Yt(e, r), n;
}
function Oo(e, t, r, ...n) {
  try {
    return r(...n);
  } catch (i) {
    Fe(i, e, t);
  }
}
function Fe(e, t, r = void 0) {
  e = Object.assign(
    e ?? { message: "No error message given." },
    { el: t, expression: r }
  ), console.warn(`Alpine Expression Error: ${e.message}

${r ? 'Expression: "' + r + `"

` : ""}`, t), setTimeout(() => {
    throw e;
  }, 0);
}
var He = !0;
function gn(e) {
  let t = He;
  He = !1;
  let r = e();
  return He = t, r;
}
function ce(e, t, r = {}) {
  let n;
  return j(e, t)((i) => n = i, r), n;
}
function j(...e) {
  return mn(...e);
}
var mn = bn;
function To(e) {
  mn = e;
}
function bn(e, t) {
  let r = {};
  Mt(r, e);
  let n = [r, ...ge(e)], i = typeof t == "function" ? $o(n, t) : Mo(n, t, e);
  return Oo.bind(null, e, t, i);
}
function $o(e, t) {
  return (r = () => {
  }, { scope: n = {}, params: i = [] } = {}) => {
    let o = t.apply(ze([n, ...e]), i);
    Xe(r, o);
  };
}
var gt = {};
function Ro(e, t) {
  if (gt[e])
    return gt[e];
  let r = Object.getPrototypeOf(async function() {
  }).constructor, n = /^[\n\s]*if.*\(.*\)/.test(e.trim()) || /^(let|const)\s/.test(e.trim()) ? `(async()=>{ ${e} })()` : e, o = (() => {
    try {
      let a = new r(
        ["__self", "scope"],
        `with (scope) { __self.result = ${n} }; __self.finished = true; return __self.result;`
      );
      return Object.defineProperty(a, "name", {
        value: `[Alpine] ${e}`
      }), a;
    } catch (a) {
      return Fe(a, t, e), Promise.resolve();
    }
  })();
  return gt[e] = o, o;
}
function Mo(e, t, r) {
  let n = Ro(t, r);
  return (i = () => {
  }, { scope: o = {}, params: a = [] } = {}) => {
    n.result = void 0, n.finished = !1;
    let s = ze([o, ...e]);
    if (typeof n == "function") {
      let l = n(n, s).catch((c) => Fe(c, r, t));
      n.finished ? (Xe(i, n.result, s, a, r), n.result = void 0) : l.then((c) => {
        Xe(i, c, s, a, r);
      }).catch((c) => Fe(c, r, t)).finally(() => n.result = void 0);
    }
  };
}
function Xe(e, t, r, n, i) {
  if (He && typeof t == "function") {
    let o = t.apply(r, n);
    o instanceof Promise ? o.then((a) => Xe(e, a, r, n)).catch((a) => Fe(a, i, t)) : e(o);
  } else typeof t == "object" && t instanceof Promise ? t.then((o) => e(o)) : e(t);
}
var rr = "x-";
function xe(e = "") {
  return rr + e;
}
function jo(e) {
  rr = e;
}
var Ze = {};
function O(e, t) {
  return Ze[e] = t, {
    before(r) {
      if (!Ze[r]) {
        console.warn(String.raw`Cannot find directive \`${r}\`. \`${e}\` will use the default order of execution`);
        return;
      }
      const n = se.indexOf(r);
      se.splice(n >= 0 ? n : se.indexOf("DEFAULT"), 0, e);
    }
  };
}
function Fo(e) {
  return Object.keys(Ze).includes(e);
}
function nr(e, t, r) {
  if (t = Array.from(t), e._x_virtualDirectives) {
    let o = Object.entries(e._x_virtualDirectives).map(([s, l]) => ({ name: s, value: l })), a = yn(o);
    o = o.map((s) => a.find((l) => l.name === s.name) ? {
      name: `x-bind:${s.name}`,
      value: `"${s.value}"`
    } : s), t = t.concat(o);
  }
  let n = {};
  return t.map(Sn((o, a) => n[o] = a)).filter(An).map(Lo(n, r)).sort(Do).map((o) => zo(e, o));
}
function yn(e) {
  return Array.from(e).map(Sn()).filter((t) => !An(t));
}
var jt = !1, $e = /* @__PURE__ */ new Map(), xn = Symbol();
function Io(e) {
  jt = !0;
  let t = Symbol();
  xn = t, $e.set(t, []);
  let r = () => {
    for (; $e.get(t).length; )
      $e.get(t).shift()();
    $e.delete(t);
  }, n = () => {
    jt = !1, r();
  };
  e(r), n();
}
function wn(e) {
  let t = [], r = (s) => t.push(s), [n, i] = yo(e);
  return t.push(i), [{
    Alpine: Le,
    effect: n,
    cleanup: r,
    evaluateLater: j.bind(j, e),
    evaluate: ce.bind(ce, e)
  }, () => t.forEach((s) => s())];
}
function zo(e, t) {
  let r = () => {
  }, n = Ze[t.type] || r, [i, o] = wn(e);
  ln(e, t.original, o);
  let a = () => {
    e._x_ignore || e._x_ignoreSelf || (n.inline && n.inline(e, t, i), n = n.bind(n, e, t, i), jt ? $e.get(xn).push(n) : n());
  };
  return a.runCleanups = o, a;
}
var _n = (e, t) => ({ name: r, value: n }) => (r.startsWith(e) && (r = r.replace(e, t)), { name: r, value: n }), vn = (e) => e;
function Sn(e = () => {
}) {
  return ({ name: t, value: r }) => {
    let { name: n, value: i } = En.reduce((o, a) => a(o), { name: t, value: r });
    return n !== t && e(n, t), { name: n, value: i };
  };
}
var En = [];
function ir(e) {
  En.push(e);
}
function An({ name: e }) {
  return Cn().test(e);
}
var Cn = () => new RegExp(`^${rr}([^:^.]+)\\b`);
function Lo(e, t) {
  return ({ name: r, value: n }) => {
    let i = r.match(Cn()), o = r.match(/:([a-zA-Z0-9\-_:]+)/), a = r.match(/\.[^.\]]+(?=[^\]]*$)/g) || [], s = t || e[r] || r;
    return {
      type: i ? i[1] : null,
      value: o ? o[1] : null,
      modifiers: a.map((l) => l.replace(".", "")),
      expression: n,
      original: s
    };
  };
}
var Ft = "DEFAULT", se = [
  "ignore",
  "ref",
  "data",
  "id",
  "anchor",
  "bind",
  "init",
  "for",
  "model",
  "modelable",
  "transition",
  "show",
  "if",
  Ft,
  "teleport"
];
function Do(e, t) {
  let r = se.indexOf(e.type) === -1 ? Ft : e.type, n = se.indexOf(t.type) === -1 ? Ft : t.type;
  return se.indexOf(r) - se.indexOf(n);
}
function Re(e, t, r = {}) {
  e.dispatchEvent(
    new CustomEvent(t, {
      detail: r,
      bubbles: !0,
      // Allows events to pass the shadow DOM barrier.
      composed: !0,
      cancelable: !0
    })
  );
}
function de(e, t) {
  if (typeof ShadowRoot == "function" && e instanceof ShadowRoot) {
    Array.from(e.children).forEach((i) => de(i, t));
    return;
  }
  let r = !1;
  if (t(e, () => r = !0), r)
    return;
  let n = e.firstElementChild;
  for (; n; )
    de(n, t), n = n.nextElementSibling;
}
function W(e, ...t) {
  console.warn(`Alpine Warning: ${e}`, ...t);
}
var $r = !1;
function Po() {
  $r && W("Alpine has already been initialized on this page. Calling Alpine.start() more than once can cause problems."), $r = !0, document.body || W("Unable to initialize. Trying to load Alpine before `<body>` is available. Did you forget to add `defer` in Alpine's `<script>` tag?"), Re(document, "alpine:init"), Re(document, "alpine:initializing"), Qt(), xo((t) => G(t, de)), Yt((t) => _e(t)), sn((t, r) => {
    nr(t, r).forEach((n) => n());
  });
  let e = (t) => !et(t.parentElement, !0);
  Array.from(document.querySelectorAll(Tn().join(","))).filter(e).forEach((t) => {
    G(t);
  }), Re(document, "alpine:initialized"), setTimeout(() => {
    Vo();
  });
}
var or = [], kn = [];
function On() {
  return or.map((e) => e());
}
function Tn() {
  return or.concat(kn).map((e) => e());
}
function $n(e) {
  or.push(e);
}
function Rn(e) {
  kn.push(e);
}
function et(e, t = !1) {
  return we(e, (r) => {
    if ((t ? Tn() : On()).some((i) => r.matches(i)))
      return !0;
  });
}
function we(e, t) {
  if (e) {
    if (t(e))
      return e;
    if (e._x_teleportBack && (e = e._x_teleportBack), !!e.parentElement)
      return we(e.parentElement, t);
  }
}
function No(e) {
  return On().some((t) => e.matches(t));
}
var Mn = [];
function Wo(e) {
  Mn.push(e);
}
var Bo = 1;
function G(e, t = de, r = () => {
}) {
  we(e, (n) => n._x_ignore) || Io(() => {
    t(e, (n, i) => {
      n._x_marker || (r(n, i), Mn.forEach((o) => o(n, i)), nr(n, n.attributes).forEach((o) => o()), n._x_ignore || (n._x_marker = Bo++), n._x_ignore && i());
    });
  });
}
function _e(e, t = de) {
  t(e, (r) => {
    wo(r), cn(r), delete r._x_marker;
  });
}
function Vo() {
  [
    ["ui", "dialog", ["[x-dialog], [x-popover]"]],
    ["anchor", "anchor", ["[x-anchor]"]],
    ["sort", "sort", ["[x-sort]"]]
  ].forEach(([t, r, n]) => {
    Fo(r) || n.some((i) => {
      if (document.querySelector(i))
        return W(`found "${i}", but missing ${t} plugin`), !0;
    });
  });
}
var It = [], ar = !1;
function sr(e = () => {
}) {
  return queueMicrotask(() => {
    ar || setTimeout(() => {
      zt();
    });
  }), new Promise((t) => {
    It.push(() => {
      e(), t();
    });
  });
}
function zt() {
  for (ar = !1; It.length; )
    It.shift()();
}
function Uo() {
  ar = !0;
}
function lr(e, t) {
  return Array.isArray(t) ? Rr(e, t.join(" ")) : typeof t == "object" && t !== null ? Ho(e, t) : typeof t == "function" ? lr(e, t()) : Rr(e, t);
}
function Rr(e, t) {
  let r = (i) => i.split(" ").filter((o) => !e.classList.contains(o)).filter(Boolean), n = (i) => (e.classList.add(...i), () => {
    e.classList.remove(...i);
  });
  return t = t === !0 ? t = "" : t || "", n(r(t));
}
function Ho(e, t) {
  let r = (s) => s.split(" ").filter(Boolean), n = Object.entries(t).flatMap(([s, l]) => l ? r(s) : !1).filter(Boolean), i = Object.entries(t).flatMap(([s, l]) => l ? !1 : r(s)).filter(Boolean), o = [], a = [];
  return i.forEach((s) => {
    e.classList.contains(s) && (e.classList.remove(s), a.push(s));
  }), n.forEach((s) => {
    e.classList.contains(s) || (e.classList.add(s), o.push(s));
  }), () => {
    a.forEach((s) => e.classList.add(s)), o.forEach((s) => e.classList.remove(s));
  };
}
function tt(e, t) {
  return typeof t == "object" && t !== null ? qo(e, t) : Ko(e, t);
}
function qo(e, t) {
  let r = {};
  return Object.entries(t).forEach(([n, i]) => {
    r[n] = e.style[n], n.startsWith("--") || (n = Jo(n)), e.style.setProperty(n, i);
  }), setTimeout(() => {
    e.style.length === 0 && e.removeAttribute("style");
  }), () => {
    tt(e, r);
  };
}
function Ko(e, t) {
  let r = e.getAttribute("style", t);
  return e.setAttribute("style", t), () => {
    e.setAttribute("style", r || "");
  };
}
function Jo(e) {
  return e.replace(/([a-z])([A-Z])/g, "$1-$2").toLowerCase();
}
function Lt(e, t = () => {
}) {
  let r = !1;
  return function() {
    r ? t.apply(this, arguments) : (r = !0, e.apply(this, arguments));
  };
}
O("transition", (e, { value: t, modifiers: r, expression: n }, { evaluate: i }) => {
  typeof n == "function" && (n = i(n)), n !== !1 && (!n || typeof n == "boolean" ? Yo(e, r, t) : Go(e, n, t));
});
function Go(e, t, r) {
  jn(e, lr, ""), {
    enter: (i) => {
      e._x_transition.enter.during = i;
    },
    "enter-start": (i) => {
      e._x_transition.enter.start = i;
    },
    "enter-end": (i) => {
      e._x_transition.enter.end = i;
    },
    leave: (i) => {
      e._x_transition.leave.during = i;
    },
    "leave-start": (i) => {
      e._x_transition.leave.start = i;
    },
    "leave-end": (i) => {
      e._x_transition.leave.end = i;
    }
  }[r](t);
}
function Yo(e, t, r) {
  jn(e, tt);
  let n = !t.includes("in") && !t.includes("out") && !r, i = n || t.includes("in") || ["enter"].includes(r), o = n || t.includes("out") || ["leave"].includes(r);
  t.includes("in") && !n && (t = t.filter((w, m) => m < t.indexOf("out"))), t.includes("out") && !n && (t = t.filter((w, m) => m > t.indexOf("out")));
  let a = !t.includes("opacity") && !t.includes("scale"), s = a || t.includes("opacity"), l = a || t.includes("scale"), c = s ? 0 : 1, f = l ? Oe(t, "scale", 95) / 100 : 1, u = Oe(t, "delay", 0) / 1e3, h = Oe(t, "origin", "center"), b = "opacity, transform", _ = Oe(t, "duration", 150) / 1e3, C = Oe(t, "duration", 75) / 1e3, g = "cubic-bezier(0.4, 0.0, 0.2, 1)";
  i && (e._x_transition.enter.during = {
    transformOrigin: h,
    transitionDelay: `${u}s`,
    transitionProperty: b,
    transitionDuration: `${_}s`,
    transitionTimingFunction: g
  }, e._x_transition.enter.start = {
    opacity: c,
    transform: `scale(${f})`
  }, e._x_transition.enter.end = {
    opacity: 1,
    transform: "scale(1)"
  }), o && (e._x_transition.leave.during = {
    transformOrigin: h,
    transitionDelay: `${u}s`,
    transitionProperty: b,
    transitionDuration: `${C}s`,
    transitionTimingFunction: g
  }, e._x_transition.leave.start = {
    opacity: 1,
    transform: "scale(1)"
  }, e._x_transition.leave.end = {
    opacity: c,
    transform: `scale(${f})`
  });
}
function jn(e, t, r = {}) {
  e._x_transition || (e._x_transition = {
    enter: { during: r, start: r, end: r },
    leave: { during: r, start: r, end: r },
    in(n = () => {
    }, i = () => {
    }) {
      Dt(e, t, {
        during: this.enter.during,
        start: this.enter.start,
        end: this.enter.end
      }, n, i);
    },
    out(n = () => {
    }, i = () => {
    }) {
      Dt(e, t, {
        during: this.leave.during,
        start: this.leave.start,
        end: this.leave.end
      }, n, i);
    }
  });
}
window.Element.prototype._x_toggleAndCascadeWithTransitions = function(e, t, r, n) {
  const i = document.visibilityState === "visible" ? requestAnimationFrame : setTimeout;
  let o = () => i(r);
  if (t) {
    e._x_transition && (e._x_transition.enter || e._x_transition.leave) ? e._x_transition.enter && (Object.entries(e._x_transition.enter.during).length || Object.entries(e._x_transition.enter.start).length || Object.entries(e._x_transition.enter.end).length) ? e._x_transition.in(r) : o() : e._x_transition ? e._x_transition.in(r) : o();
    return;
  }
  e._x_hidePromise = e._x_transition ? new Promise((a, s) => {
    e._x_transition.out(() => {
    }, () => a(n)), e._x_transitioning && e._x_transitioning.beforeCancel(() => s({ isFromCancelledTransition: !0 }));
  }) : Promise.resolve(n), queueMicrotask(() => {
    let a = Fn(e);
    a ? (a._x_hideChildren || (a._x_hideChildren = []), a._x_hideChildren.push(e)) : i(() => {
      let s = (l) => {
        let c = Promise.all([
          l._x_hidePromise,
          ...(l._x_hideChildren || []).map(s)
        ]).then(([f]) => f == null ? void 0 : f());
        return delete l._x_hidePromise, delete l._x_hideChildren, c;
      };
      s(e).catch((l) => {
        if (!l.isFromCancelledTransition)
          throw l;
      });
    });
  });
};
function Fn(e) {
  let t = e.parentNode;
  if (t)
    return t._x_hidePromise ? t : Fn(t);
}
function Dt(e, t, { during: r, start: n, end: i } = {}, o = () => {
}, a = () => {
}) {
  if (e._x_transitioning && e._x_transitioning.cancel(), Object.keys(r).length === 0 && Object.keys(n).length === 0 && Object.keys(i).length === 0) {
    o(), a();
    return;
  }
  let s, l, c;
  Xo(e, {
    start() {
      s = t(e, n);
    },
    during() {
      l = t(e, r);
    },
    before: o,
    end() {
      s(), c = t(e, i);
    },
    after: a,
    cleanup() {
      l(), c();
    }
  });
}
function Xo(e, t) {
  let r, n, i, o = Lt(() => {
    k(() => {
      r = !0, n || t.before(), i || (t.end(), zt()), t.after(), e.isConnected && t.cleanup(), delete e._x_transitioning;
    });
  });
  e._x_transitioning = {
    beforeCancels: [],
    beforeCancel(a) {
      this.beforeCancels.push(a);
    },
    cancel: Lt(function() {
      for (; this.beforeCancels.length; )
        this.beforeCancels.shift()();
      o();
    }),
    finish: o
  }, k(() => {
    t.start(), t.during();
  }), Uo(), requestAnimationFrame(() => {
    if (r)
      return;
    let a = Number(getComputedStyle(e).transitionDuration.replace(/,.*/, "").replace("s", "")) * 1e3, s = Number(getComputedStyle(e).transitionDelay.replace(/,.*/, "").replace("s", "")) * 1e3;
    a === 0 && (a = Number(getComputedStyle(e).animationDuration.replace("s", "")) * 1e3), k(() => {
      t.before();
    }), n = !0, requestAnimationFrame(() => {
      r || (k(() => {
        t.end();
      }), zt(), setTimeout(e._x_transitioning.finish, a + s), i = !0);
    });
  });
}
function Oe(e, t, r) {
  if (e.indexOf(t) === -1)
    return r;
  const n = e[e.indexOf(t) + 1];
  if (!n || t === "scale" && isNaN(n))
    return r;
  if (t === "duration" || t === "delay") {
    let i = n.match(/([0-9]+)ms/);
    if (i)
      return i[1];
  }
  return t === "origin" && ["top", "right", "left", "center", "bottom"].includes(e[e.indexOf(t) + 2]) ? [n, e[e.indexOf(t) + 2]].join(" ") : n;
}
var ee = !1;
function re(e, t = () => {
}) {
  return (...r) => ee ? t(...r) : e(...r);
}
function Zo(e) {
  return (...t) => ee && e(...t);
}
var In = [];
function rt(e) {
  In.push(e);
}
function Qo(e, t) {
  In.forEach((r) => r(e, t)), ee = !0, zn(() => {
    G(t, (r, n) => {
      n(r, () => {
      });
    });
  }), ee = !1;
}
var Pt = !1;
function ea(e, t) {
  t._x_dataStack || (t._x_dataStack = e._x_dataStack), ee = !0, Pt = !0, zn(() => {
    ta(t);
  }), ee = !1, Pt = !1;
}
function ta(e) {
  let t = !1;
  G(e, (n, i) => {
    de(n, (o, a) => {
      if (t && No(o))
        return a();
      t = !0, i(o, a);
    });
  });
}
function zn(e) {
  let t = pe;
  Tr((r, n) => {
    let i = t(r);
    return ye(i), () => {
    };
  }), e(), Tr(t);
}
function Ln(e, t, r, n = []) {
  switch (e._x_bindings || (e._x_bindings = be({})), e._x_bindings[t] = r, t = n.includes("camel") ? ca(t) : t, t) {
    case "value":
      ra(e, r);
      break;
    case "style":
      ia(e, r);
      break;
    case "class":
      na(e, r);
      break;
    case "selected":
    case "checked":
      oa(e, t, r);
      break;
    default:
      Dn(e, t, r);
      break;
  }
}
function ra(e, t) {
  if (Wn(e))
    e.attributes.value === void 0 && (e.value = t), window.fromModel && (typeof t == "boolean" ? e.checked = qe(e.value) === t : e.checked = Mr(e.value, t));
  else if (cr(e))
    Number.isInteger(t) ? e.value = t : !Array.isArray(t) && typeof t != "boolean" && ![null, void 0].includes(t) ? e.value = String(t) : Array.isArray(t) ? e.checked = t.some((r) => Mr(r, e.value)) : e.checked = !!t;
  else if (e.tagName === "SELECT")
    la(e, t);
  else {
    if (e.value === t)
      return;
    e.value = t === void 0 ? "" : t;
  }
}
function na(e, t) {
  e._x_undoAddedClasses && e._x_undoAddedClasses(), e._x_undoAddedClasses = lr(e, t);
}
function ia(e, t) {
  e._x_undoAddedStyles && e._x_undoAddedStyles(), e._x_undoAddedStyles = tt(e, t);
}
function oa(e, t, r) {
  Dn(e, t, r), sa(e, t, r);
}
function Dn(e, t, r) {
  [null, void 0, !1].includes(r) && ua(t) ? e.removeAttribute(t) : (Pn(t) && (r = t), aa(e, t, r));
}
function aa(e, t, r) {
  e.getAttribute(t) != r && e.setAttribute(t, r);
}
function sa(e, t, r) {
  e[t] !== r && (e[t] = r);
}
function la(e, t) {
  const r = [].concat(t).map((n) => n + "");
  Array.from(e.options).forEach((n) => {
    n.selected = r.includes(n.value);
  });
}
function ca(e) {
  return e.toLowerCase().replace(/-(\w)/g, (t, r) => r.toUpperCase());
}
function Mr(e, t) {
  return e == t;
}
function qe(e) {
  return [1, "1", "true", "on", "yes", !0].includes(e) ? !0 : [0, "0", "false", "off", "no", !1].includes(e) ? !1 : e ? !!e : null;
}
var fa = /* @__PURE__ */ new Set([
  "allowfullscreen",
  "async",
  "autofocus",
  "autoplay",
  "checked",
  "controls",
  "default",
  "defer",
  "disabled",
  "formnovalidate",
  "inert",
  "ismap",
  "itemscope",
  "loop",
  "multiple",
  "muted",
  "nomodule",
  "novalidate",
  "open",
  "playsinline",
  "readonly",
  "required",
  "reversed",
  "selected",
  "shadowrootclonable",
  "shadowrootdelegatesfocus",
  "shadowrootserializable"
]);
function Pn(e) {
  return fa.has(e);
}
function ua(e) {
  return !["aria-pressed", "aria-checked", "aria-expanded", "aria-selected"].includes(e);
}
function da(e, t, r) {
  return e._x_bindings && e._x_bindings[t] !== void 0 ? e._x_bindings[t] : Nn(e, t, r);
}
function pa(e, t, r, n = !0) {
  if (e._x_bindings && e._x_bindings[t] !== void 0)
    return e._x_bindings[t];
  if (e._x_inlineBindings && e._x_inlineBindings[t] !== void 0) {
    let i = e._x_inlineBindings[t];
    return i.extract = n, gn(() => ce(e, i.expression));
  }
  return Nn(e, t, r);
}
function Nn(e, t, r) {
  let n = e.getAttribute(t);
  return n === null ? typeof r == "function" ? r() : r : n === "" ? !0 : Pn(t) ? !![t, "true"].includes(n) : n;
}
function cr(e) {
  return e.type === "checkbox" || e.localName === "ui-checkbox" || e.localName === "ui-switch";
}
function Wn(e) {
  return e.type === "radio" || e.localName === "ui-radio";
}
function Bn(e, t) {
  var r;
  return function() {
    var n = this, i = arguments, o = function() {
      r = null, e.apply(n, i);
    };
    clearTimeout(r), r = setTimeout(o, t);
  };
}
function Vn(e, t) {
  let r;
  return function() {
    let n = this, i = arguments;
    r || (e.apply(n, i), r = !0, setTimeout(() => r = !1, t));
  };
}
function Un({ get: e, set: t }, { get: r, set: n }) {
  let i = !0, o, a = pe(() => {
    let s = e(), l = r();
    if (i)
      n(mt(s)), i = !1;
    else {
      let c = JSON.stringify(s), f = JSON.stringify(l);
      c !== o ? n(mt(s)) : c !== f && t(mt(l));
    }
    o = JSON.stringify(e()), JSON.stringify(r());
  });
  return () => {
    ye(a);
  };
}
function mt(e) {
  return typeof e == "object" ? JSON.parse(JSON.stringify(e)) : e;
}
function ha(e) {
  (Array.isArray(e) ? e : [e]).forEach((r) => r(Le));
}
var oe = {}, jr = !1;
function ga(e, t) {
  if (jr || (oe = be(oe), jr = !0), t === void 0)
    return oe[e];
  oe[e] = t, dn(oe[e]), typeof t == "object" && t !== null && t.hasOwnProperty("init") && typeof t.init == "function" && oe[e].init();
}
function ma() {
  return oe;
}
var Hn = {};
function ba(e, t) {
  let r = typeof t != "function" ? () => t : t;
  return e instanceof Element ? qn(e, r()) : (Hn[e] = r, () => {
  });
}
function ya(e) {
  return Object.entries(Hn).forEach(([t, r]) => {
    Object.defineProperty(e, t, {
      get() {
        return (...n) => r(...n);
      }
    });
  }), e;
}
function qn(e, t, r) {
  let n = [];
  for (; n.length; )
    n.pop()();
  let i = Object.entries(t).map(([a, s]) => ({ name: a, value: s })), o = yn(i);
  return i = i.map((a) => o.find((s) => s.name === a.name) ? {
    name: `x-bind:${a.name}`,
    value: `"${a.value}"`
  } : a), nr(e, i, r).map((a) => {
    n.push(a.runCleanups), a();
  }), () => {
    for (; n.length; )
      n.pop()();
  };
}
var Kn = {};
function xa(e, t) {
  Kn[e] = t;
}
function wa(e, t) {
  return Object.entries(Kn).forEach(([r, n]) => {
    Object.defineProperty(e, r, {
      get() {
        return (...i) => n.bind(t)(...i);
      },
      enumerable: !1
    });
  }), e;
}
var _a = {
  get reactive() {
    return be;
  },
  get release() {
    return ye;
  },
  get effect() {
    return pe;
  },
  get raw() {
    return tn;
  },
  version: "3.14.8",
  flushAndStopDeferringMutations: So,
  dontAutoEvaluateFunctions: gn,
  disableEffectScheduling: mo,
  startObservingMutations: Qt,
  stopObservingMutations: fn,
  setReactivityEngine: bo,
  onAttributeRemoved: ln,
  onAttributesAdded: sn,
  closestDataStack: ge,
  skipDuringClone: re,
  onlyDuringClone: Zo,
  addRootSelector: $n,
  addInitSelector: Rn,
  interceptClone: rt,
  addScopeToNode: Ie,
  deferMutations: vo,
  mapAttributes: ir,
  evaluateLater: j,
  interceptInit: Wo,
  setEvaluator: To,
  mergeProxies: ze,
  extractProp: pa,
  findClosest: we,
  onElRemoved: Yt,
  closestRoot: et,
  destroyTree: _e,
  interceptor: pn,
  // INTERNAL: not public API and is subject to change without major release.
  transition: Dt,
  // INTERNAL
  setStyles: tt,
  // INTERNAL
  mutateDom: k,
  directive: O,
  entangle: Un,
  throttle: Vn,
  debounce: Bn,
  evaluate: ce,
  initTree: G,
  nextTick: sr,
  prefixed: xe,
  prefix: jo,
  plugin: ha,
  magic: H,
  store: ga,
  start: Po,
  clone: ea,
  // INTERNAL
  cloneNode: Qo,
  // INTERNAL
  bound: da,
  $data: un,
  watch: rn,
  walk: de,
  data: xa,
  bind: ba
}, Le = _a;
function va(e, t) {
  const r = /* @__PURE__ */ Object.create(null), n = e.split(",");
  for (let i = 0; i < n.length; i++)
    r[n[i]] = !0;
  return (i) => !!r[i];
}
var Sa = Object.freeze({}), Ea = Object.prototype.hasOwnProperty, nt = (e, t) => Ea.call(e, t), fe = Array.isArray, Me = (e) => Jn(e) === "[object Map]", Aa = (e) => typeof e == "string", fr = (e) => typeof e == "symbol", it = (e) => e !== null && typeof e == "object", Ca = Object.prototype.toString, Jn = (e) => Ca.call(e), Gn = (e) => Jn(e).slice(8, -1), ur = (e) => Aa(e) && e !== "NaN" && e[0] !== "-" && "" + parseInt(e, 10) === e, ka = (e) => {
  const t = /* @__PURE__ */ Object.create(null);
  return (r) => t[r] || (t[r] = e(r));
}, Oa = ka((e) => e.charAt(0).toUpperCase() + e.slice(1)), Yn = (e, t) => e !== t && (e === e || t === t), Nt = /* @__PURE__ */ new WeakMap(), Te = [], J, ue = Symbol("iterate"), Wt = Symbol("Map key iterate");
function Ta(e) {
  return e && e._isEffect === !0;
}
function $a(e, t = Sa) {
  Ta(e) && (e = e.raw);
  const r = ja(e, t);
  return t.lazy || r(), r;
}
function Ra(e) {
  e.active && (Xn(e), e.options.onStop && e.options.onStop(), e.active = !1);
}
var Ma = 0;
function ja(e, t) {
  const r = function() {
    if (!r.active)
      return e();
    if (!Te.includes(r)) {
      Xn(r);
      try {
        return Ia(), Te.push(r), J = r, e();
      } finally {
        Te.pop(), Zn(), J = Te[Te.length - 1];
      }
    }
  };
  return r.id = Ma++, r.allowRecurse = !!t.allowRecurse, r._isEffect = !0, r.active = !0, r.raw = e, r.deps = [], r.options = t, r;
}
function Xn(e) {
  const { deps: t } = e;
  if (t.length) {
    for (let r = 0; r < t.length; r++)
      t[r].delete(e);
    t.length = 0;
  }
}
var me = !0, dr = [];
function Fa() {
  dr.push(me), me = !1;
}
function Ia() {
  dr.push(me), me = !0;
}
function Zn() {
  const e = dr.pop();
  me = e === void 0 ? !0 : e;
}
function U(e, t, r) {
  if (!me || J === void 0)
    return;
  let n = Nt.get(e);
  n || Nt.set(e, n = /* @__PURE__ */ new Map());
  let i = n.get(r);
  i || n.set(r, i = /* @__PURE__ */ new Set()), i.has(J) || (i.add(J), J.deps.push(i), J.options.onTrack && J.options.onTrack({
    effect: J,
    target: e,
    type: t,
    key: r
  }));
}
function te(e, t, r, n, i, o) {
  const a = Nt.get(e);
  if (!a)
    return;
  const s = /* @__PURE__ */ new Set(), l = (f) => {
    f && f.forEach((u) => {
      (u !== J || u.allowRecurse) && s.add(u);
    });
  };
  if (t === "clear")
    a.forEach(l);
  else if (r === "length" && fe(e))
    a.forEach((f, u) => {
      (u === "length" || u >= n) && l(f);
    });
  else
    switch (r !== void 0 && l(a.get(r)), t) {
      case "add":
        fe(e) ? ur(r) && l(a.get("length")) : (l(a.get(ue)), Me(e) && l(a.get(Wt)));
        break;
      case "delete":
        fe(e) || (l(a.get(ue)), Me(e) && l(a.get(Wt)));
        break;
      case "set":
        Me(e) && l(a.get(ue));
        break;
    }
  const c = (f) => {
    f.options.onTrigger && f.options.onTrigger({
      effect: f,
      target: e,
      key: r,
      type: t,
      newValue: n,
      oldValue: i,
      oldTarget: o
    }), f.options.scheduler ? f.options.scheduler(f) : f();
  };
  s.forEach(c);
}
var za = /* @__PURE__ */ va("__proto__,__v_isRef,__isVue"), Qn = new Set(Object.getOwnPropertyNames(Symbol).map((e) => Symbol[e]).filter(fr)), La = /* @__PURE__ */ ei(), Da = /* @__PURE__ */ ei(!0), Fr = /* @__PURE__ */ Pa();
function Pa() {
  const e = {};
  return ["includes", "indexOf", "lastIndexOf"].forEach((t) => {
    e[t] = function(...r) {
      const n = A(this);
      for (let o = 0, a = this.length; o < a; o++)
        U(n, "get", o + "");
      const i = n[t](...r);
      return i === -1 || i === !1 ? n[t](...r.map(A)) : i;
    };
  }), ["push", "pop", "shift", "unshift", "splice"].forEach((t) => {
    e[t] = function(...r) {
      Fa();
      const n = A(this)[t].apply(this, r);
      return Zn(), n;
    };
  }), e;
}
function ei(e = !1, t = !1) {
  return function(n, i, o) {
    if (i === "__v_isReactive")
      return !e;
    if (i === "__v_isReadonly")
      return e;
    if (i === "__v_raw" && o === (e ? t ? Qa : ii : t ? Za : ni).get(n))
      return n;
    const a = fe(n);
    if (!e && a && nt(Fr, i))
      return Reflect.get(Fr, i, o);
    const s = Reflect.get(n, i, o);
    return (fr(i) ? Qn.has(i) : za(i)) || (e || U(n, "get", i), t) ? s : Bt(s) ? !a || !ur(i) ? s.value : s : it(s) ? e ? oi(s) : mr(s) : s;
  };
}
var Na = /* @__PURE__ */ Wa();
function Wa(e = !1) {
  return function(r, n, i, o) {
    let a = r[n];
    if (!e && (i = A(i), a = A(a), !fe(r) && Bt(a) && !Bt(i)))
      return a.value = i, !0;
    const s = fe(r) && ur(n) ? Number(n) < r.length : nt(r, n), l = Reflect.set(r, n, i, o);
    return r === A(o) && (s ? Yn(i, a) && te(r, "set", n, i, a) : te(r, "add", n, i)), l;
  };
}
function Ba(e, t) {
  const r = nt(e, t), n = e[t], i = Reflect.deleteProperty(e, t);
  return i && r && te(e, "delete", t, void 0, n), i;
}
function Va(e, t) {
  const r = Reflect.has(e, t);
  return (!fr(t) || !Qn.has(t)) && U(e, "has", t), r;
}
function Ua(e) {
  return U(e, "iterate", fe(e) ? "length" : ue), Reflect.ownKeys(e);
}
var Ha = {
  get: La,
  set: Na,
  deleteProperty: Ba,
  has: Va,
  ownKeys: Ua
}, qa = {
  get: Da,
  set(e, t) {
    return console.warn(`Set operation on key "${String(t)}" failed: target is readonly.`, e), !0;
  },
  deleteProperty(e, t) {
    return console.warn(`Delete operation on key "${String(t)}" failed: target is readonly.`, e), !0;
  }
}, pr = (e) => it(e) ? mr(e) : e, hr = (e) => it(e) ? oi(e) : e, gr = (e) => e, ot = (e) => Reflect.getPrototypeOf(e);
function Pe(e, t, r = !1, n = !1) {
  e = e.__v_raw;
  const i = A(e), o = A(t);
  t !== o && !r && U(i, "get", t), !r && U(i, "get", o);
  const { has: a } = ot(i), s = n ? gr : r ? hr : pr;
  if (a.call(i, t))
    return s(e.get(t));
  if (a.call(i, o))
    return s(e.get(o));
  e !== i && e.get(t);
}
function Ne(e, t = !1) {
  const r = this.__v_raw, n = A(r), i = A(e);
  return e !== i && !t && U(n, "has", e), !t && U(n, "has", i), e === i ? r.has(e) : r.has(e) || r.has(i);
}
function We(e, t = !1) {
  return e = e.__v_raw, !t && U(A(e), "iterate", ue), Reflect.get(e, "size", e);
}
function Ir(e) {
  e = A(e);
  const t = A(this);
  return ot(t).has.call(t, e) || (t.add(e), te(t, "add", e, e)), this;
}
function zr(e, t) {
  t = A(t);
  const r = A(this), { has: n, get: i } = ot(r);
  let o = n.call(r, e);
  o ? ri(r, n, e) : (e = A(e), o = n.call(r, e));
  const a = i.call(r, e);
  return r.set(e, t), o ? Yn(t, a) && te(r, "set", e, t, a) : te(r, "add", e, t), this;
}
function Lr(e) {
  const t = A(this), { has: r, get: n } = ot(t);
  let i = r.call(t, e);
  i ? ri(t, r, e) : (e = A(e), i = r.call(t, e));
  const o = n ? n.call(t, e) : void 0, a = t.delete(e);
  return i && te(t, "delete", e, void 0, o), a;
}
function Dr() {
  const e = A(this), t = e.size !== 0, r = Me(e) ? new Map(e) : new Set(e), n = e.clear();
  return t && te(e, "clear", void 0, void 0, r), n;
}
function Be(e, t) {
  return function(n, i) {
    const o = this, a = o.__v_raw, s = A(a), l = t ? gr : e ? hr : pr;
    return !e && U(s, "iterate", ue), a.forEach((c, f) => n.call(i, l(c), l(f), o));
  };
}
function Ve(e, t, r) {
  return function(...n) {
    const i = this.__v_raw, o = A(i), a = Me(o), s = e === "entries" || e === Symbol.iterator && a, l = e === "keys" && a, c = i[e](...n), f = r ? gr : t ? hr : pr;
    return !t && U(o, "iterate", l ? Wt : ue), {
      // iterator protocol
      next() {
        const { value: u, done: h } = c.next();
        return h ? { value: u, done: h } : {
          value: s ? [f(u[0]), f(u[1])] : f(u),
          done: h
        };
      },
      // iterable protocol
      [Symbol.iterator]() {
        return this;
      }
    };
  };
}
function Y(e) {
  return function(...t) {
    {
      const r = t[0] ? `on key "${t[0]}" ` : "";
      console.warn(`${Oa(e)} operation ${r}failed: target is readonly.`, A(this));
    }
    return e === "delete" ? !1 : this;
  };
}
function Ka() {
  const e = {
    get(o) {
      return Pe(this, o);
    },
    get size() {
      return We(this);
    },
    has: Ne,
    add: Ir,
    set: zr,
    delete: Lr,
    clear: Dr,
    forEach: Be(!1, !1)
  }, t = {
    get(o) {
      return Pe(this, o, !1, !0);
    },
    get size() {
      return We(this);
    },
    has: Ne,
    add: Ir,
    set: zr,
    delete: Lr,
    clear: Dr,
    forEach: Be(!1, !0)
  }, r = {
    get(o) {
      return Pe(this, o, !0);
    },
    get size() {
      return We(this, !0);
    },
    has(o) {
      return Ne.call(this, o, !0);
    },
    add: Y(
      "add"
      /* ADD */
    ),
    set: Y(
      "set"
      /* SET */
    ),
    delete: Y(
      "delete"
      /* DELETE */
    ),
    clear: Y(
      "clear"
      /* CLEAR */
    ),
    forEach: Be(!0, !1)
  }, n = {
    get(o) {
      return Pe(this, o, !0, !0);
    },
    get size() {
      return We(this, !0);
    },
    has(o) {
      return Ne.call(this, o, !0);
    },
    add: Y(
      "add"
      /* ADD */
    ),
    set: Y(
      "set"
      /* SET */
    ),
    delete: Y(
      "delete"
      /* DELETE */
    ),
    clear: Y(
      "clear"
      /* CLEAR */
    ),
    forEach: Be(!0, !0)
  };
  return ["keys", "values", "entries", Symbol.iterator].forEach((o) => {
    e[o] = Ve(o, !1, !1), r[o] = Ve(o, !0, !1), t[o] = Ve(o, !1, !0), n[o] = Ve(o, !0, !0);
  }), [
    e,
    r,
    t,
    n
  ];
}
var [Ja, Ga, _s, vs] = /* @__PURE__ */ Ka();
function ti(e, t) {
  const r = e ? Ga : Ja;
  return (n, i, o) => i === "__v_isReactive" ? !e : i === "__v_isReadonly" ? e : i === "__v_raw" ? n : Reflect.get(nt(r, i) && i in n ? r : n, i, o);
}
var Ya = {
  get: /* @__PURE__ */ ti(!1)
}, Xa = {
  get: /* @__PURE__ */ ti(!0)
};
function ri(e, t, r) {
  const n = A(r);
  if (n !== r && t.call(e, n)) {
    const i = Gn(e);
    console.warn(`Reactive ${i} contains both the raw and reactive versions of the same object${i === "Map" ? " as keys" : ""}, which can lead to inconsistencies. Avoid differentiating between the raw and reactive versions of an object and only use the reactive version if possible.`);
  }
}
var ni = /* @__PURE__ */ new WeakMap(), Za = /* @__PURE__ */ new WeakMap(), ii = /* @__PURE__ */ new WeakMap(), Qa = /* @__PURE__ */ new WeakMap();
function es(e) {
  switch (e) {
    case "Object":
    case "Array":
      return 1;
    case "Map":
    case "Set":
    case "WeakMap":
    case "WeakSet":
      return 2;
    default:
      return 0;
  }
}
function ts(e) {
  return e.__v_skip || !Object.isExtensible(e) ? 0 : es(Gn(e));
}
function mr(e) {
  return e && e.__v_isReadonly ? e : ai(e, !1, Ha, Ya, ni);
}
function oi(e) {
  return ai(e, !0, qa, Xa, ii);
}
function ai(e, t, r, n, i) {
  if (!it(e))
    return console.warn(`value cannot be made reactive: ${String(e)}`), e;
  if (e.__v_raw && !(t && e.__v_isReactive))
    return e;
  const o = i.get(e);
  if (o)
    return o;
  const a = ts(e);
  if (a === 0)
    return e;
  const s = new Proxy(e, a === 2 ? n : r);
  return i.set(e, s), s;
}
function A(e) {
  return e && A(e.__v_raw) || e;
}
function Bt(e) {
  return !!(e && e.__v_isRef === !0);
}
H("nextTick", () => sr);
H("dispatch", (e) => Re.bind(Re, e));
H("watch", (e, { evaluateLater: t, cleanup: r }) => (n, i) => {
  let o = t(n), s = rn(() => {
    let l;
    return o((c) => l = c), l;
  }, i);
  r(s);
});
H("store", ma);
H("data", (e) => un(e));
H("root", (e) => et(e));
H("refs", (e) => (e._x_refs_proxy || (e._x_refs_proxy = ze(rs(e))), e._x_refs_proxy));
function rs(e) {
  let t = [];
  return we(e, (r) => {
    r._x_refs && t.push(r._x_refs);
  }), t;
}
var bt = {};
function si(e) {
  return bt[e] || (bt[e] = 0), ++bt[e];
}
function ns(e, t) {
  return we(e, (r) => {
    if (r._x_ids && r._x_ids[t])
      return !0;
  });
}
function is(e, t) {
  e._x_ids || (e._x_ids = {}), e._x_ids[t] || (e._x_ids[t] = si(t));
}
H("id", (e, { cleanup: t }) => (r, n = null) => {
  let i = `${r}${n ? `-${n}` : ""}`;
  return os(e, i, t, () => {
    let o = ns(e, r), a = o ? o._x_ids[r] : si(r);
    return n ? `${r}-${a}-${n}` : `${r}-${a}`;
  });
});
rt((e, t) => {
  e._x_id && (t._x_id = e._x_id);
});
function os(e, t, r, n) {
  if (e._x_id || (e._x_id = {}), e._x_id[t])
    return e._x_id[t];
  let i = n();
  return e._x_id[t] = i, r(() => {
    delete e._x_id[t];
  }), i;
}
H("el", (e) => e);
li("Focus", "focus", "focus");
li("Persist", "persist", "persist");
function li(e, t, r) {
  H(t, (n) => W(`You can't use [$${t}] without first installing the "${e}" plugin here: https://alpinejs.dev/plugins/${r}`, n));
}
O("modelable", (e, { expression: t }, { effect: r, evaluateLater: n, cleanup: i }) => {
  let o = n(t), a = () => {
    let f;
    return o((u) => f = u), f;
  }, s = n(`${t} = __placeholder`), l = (f) => s(() => {
  }, { scope: { __placeholder: f } }), c = a();
  l(c), queueMicrotask(() => {
    if (!e._x_model)
      return;
    e._x_removeModelListeners.default();
    let f = e._x_model.get, u = e._x_model.set, h = Un(
      {
        get() {
          return f();
        },
        set(b) {
          u(b);
        }
      },
      {
        get() {
          return a();
        },
        set(b) {
          l(b);
        }
      }
    );
    i(h);
  });
});
O("teleport", (e, { modifiers: t, expression: r }, { cleanup: n }) => {
  e.tagName.toLowerCase() !== "template" && W("x-teleport can only be used on a <template> tag", e);
  let i = Pr(r), o = e.content.cloneNode(!0).firstElementChild;
  e._x_teleport = o, o._x_teleportBack = e, e.setAttribute("data-teleport-template", !0), o.setAttribute("data-teleport-target", !0), e._x_forwardEvents && e._x_forwardEvents.forEach((s) => {
    o.addEventListener(s, (l) => {
      l.stopPropagation(), e.dispatchEvent(new l.constructor(l.type, l));
    });
  }), Ie(o, {}, e);
  let a = (s, l, c) => {
    c.includes("prepend") ? l.parentNode.insertBefore(s, l) : c.includes("append") ? l.parentNode.insertBefore(s, l.nextSibling) : l.appendChild(s);
  };
  k(() => {
    a(o, i, t), re(() => {
      G(o);
    })();
  }), e._x_teleportPutBack = () => {
    let s = Pr(r);
    k(() => {
      a(e._x_teleport, s, t);
    });
  }, n(
    () => k(() => {
      o.remove(), _e(o);
    })
  );
});
var as = document.createElement("div");
function Pr(e) {
  let t = re(() => document.querySelector(e), () => as)();
  return t || W(`Cannot find x-teleport element for selector: "${e}"`), t;
}
var ci = () => {
};
ci.inline = (e, { modifiers: t }, { cleanup: r }) => {
  t.includes("self") ? e._x_ignoreSelf = !0 : e._x_ignore = !0, r(() => {
    t.includes("self") ? delete e._x_ignoreSelf : delete e._x_ignore;
  });
};
O("ignore", ci);
O("effect", re((e, { expression: t }, { effect: r }) => {
  r(j(e, t));
}));
function Vt(e, t, r, n) {
  let i = e, o = (l) => n(l), a = {}, s = (l, c) => (f) => c(l, f);
  if (r.includes("dot") && (t = ss(t)), r.includes("camel") && (t = ls(t)), r.includes("passive") && (a.passive = !0), r.includes("capture") && (a.capture = !0), r.includes("window") && (i = window), r.includes("document") && (i = document), r.includes("debounce")) {
    let l = r[r.indexOf("debounce") + 1] || "invalid-wait", c = Qe(l.split("ms")[0]) ? Number(l.split("ms")[0]) : 250;
    o = Bn(o, c);
  }
  if (r.includes("throttle")) {
    let l = r[r.indexOf("throttle") + 1] || "invalid-wait", c = Qe(l.split("ms")[0]) ? Number(l.split("ms")[0]) : 250;
    o = Vn(o, c);
  }
  return r.includes("prevent") && (o = s(o, (l, c) => {
    c.preventDefault(), l(c);
  })), r.includes("stop") && (o = s(o, (l, c) => {
    c.stopPropagation(), l(c);
  })), r.includes("once") && (o = s(o, (l, c) => {
    l(c), i.removeEventListener(t, o, a);
  })), (r.includes("away") || r.includes("outside")) && (i = document, o = s(o, (l, c) => {
    e.contains(c.target) || c.target.isConnected !== !1 && (e.offsetWidth < 1 && e.offsetHeight < 1 || e._x_isShown !== !1 && l(c));
  })), r.includes("self") && (o = s(o, (l, c) => {
    c.target === e && l(c);
  })), (fs(t) || fi(t)) && (o = s(o, (l, c) => {
    us(c, r) || l(c);
  })), i.addEventListener(t, o, a), () => {
    i.removeEventListener(t, o, a);
  };
}
function ss(e) {
  return e.replace(/-/g, ".");
}
function ls(e) {
  return e.toLowerCase().replace(/-(\w)/g, (t, r) => r.toUpperCase());
}
function Qe(e) {
  return !Array.isArray(e) && !isNaN(e);
}
function cs(e) {
  return [" ", "_"].includes(
    e
  ) ? e : e.replace(/([a-z])([A-Z])/g, "$1-$2").replace(/[_\s]/, "-").toLowerCase();
}
function fs(e) {
  return ["keydown", "keyup"].includes(e);
}
function fi(e) {
  return ["contextmenu", "click", "mouse"].some((t) => e.includes(t));
}
function us(e, t) {
  let r = t.filter((o) => !["window", "document", "prevent", "stop", "once", "capture", "self", "away", "outside", "passive"].includes(o));
  if (r.includes("debounce")) {
    let o = r.indexOf("debounce");
    r.splice(o, Qe((r[o + 1] || "invalid-wait").split("ms")[0]) ? 2 : 1);
  }
  if (r.includes("throttle")) {
    let o = r.indexOf("throttle");
    r.splice(o, Qe((r[o + 1] || "invalid-wait").split("ms")[0]) ? 2 : 1);
  }
  if (r.length === 0 || r.length === 1 && Nr(e.key).includes(r[0]))
    return !1;
  const i = ["ctrl", "shift", "alt", "meta", "cmd", "super"].filter((o) => r.includes(o));
  return r = r.filter((o) => !i.includes(o)), !(i.length > 0 && i.filter((a) => ((a === "cmd" || a === "super") && (a = "meta"), e[`${a}Key`])).length === i.length && (fi(e.type) || Nr(e.key).includes(r[0])));
}
function Nr(e) {
  if (!e)
    return [];
  e = cs(e);
  let t = {
    ctrl: "control",
    slash: "/",
    space: " ",
    spacebar: " ",
    cmd: "meta",
    esc: "escape",
    up: "arrow-up",
    down: "arrow-down",
    left: "arrow-left",
    right: "arrow-right",
    period: ".",
    comma: ",",
    equal: "=",
    minus: "-",
    underscore: "_"
  };
  return t[e] = e, Object.keys(t).map((r) => {
    if (t[r] === e)
      return r;
  }).filter((r) => r);
}
O("model", (e, { modifiers: t, expression: r }, { effect: n, cleanup: i }) => {
  let o = e;
  t.includes("parent") && (o = e.parentNode);
  let a = j(o, r), s;
  typeof r == "string" ? s = j(o, `${r} = __placeholder`) : typeof r == "function" && typeof r() == "string" ? s = j(o, `${r()} = __placeholder`) : s = () => {
  };
  let l = () => {
    let h;
    return a((b) => h = b), Wr(h) ? h.get() : h;
  }, c = (h) => {
    let b;
    a((_) => b = _), Wr(b) ? b.set(h) : s(() => {
    }, {
      scope: { __placeholder: h }
    });
  };
  typeof r == "string" && e.type === "radio" && k(() => {
    e.hasAttribute("name") || e.setAttribute("name", r);
  });
  var f = e.tagName.toLowerCase() === "select" || ["checkbox", "radio"].includes(e.type) || t.includes("lazy") ? "change" : "input";
  let u = ee ? () => {
  } : Vt(e, f, t, (h) => {
    c(yt(e, t, h, l()));
  });
  if (t.includes("fill") && ([void 0, null, ""].includes(l()) || cr(e) && Array.isArray(l()) || e.tagName.toLowerCase() === "select" && e.multiple) && c(
    yt(e, t, { target: e }, l())
  ), e._x_removeModelListeners || (e._x_removeModelListeners = {}), e._x_removeModelListeners.default = u, i(() => e._x_removeModelListeners.default()), e.form) {
    let h = Vt(e.form, "reset", [], (b) => {
      sr(() => e._x_model && e._x_model.set(yt(e, t, { target: e }, l())));
    });
    i(() => h());
  }
  e._x_model = {
    get() {
      return l();
    },
    set(h) {
      c(h);
    }
  }, e._x_forceModelUpdate = (h) => {
    h === void 0 && typeof r == "string" && r.match(/\./) && (h = ""), window.fromModel = !0, k(() => Ln(e, "value", h)), delete window.fromModel;
  }, n(() => {
    let h = l();
    t.includes("unintrusive") && document.activeElement.isSameNode(e) || e._x_forceModelUpdate(h);
  });
});
function yt(e, t, r, n) {
  return k(() => {
    if (r instanceof CustomEvent && r.detail !== void 0)
      return r.detail !== null && r.detail !== void 0 ? r.detail : r.target.value;
    if (cr(e))
      if (Array.isArray(n)) {
        let i = null;
        return t.includes("number") ? i = xt(r.target.value) : t.includes("boolean") ? i = qe(r.target.value) : i = r.target.value, r.target.checked ? n.includes(i) ? n : n.concat([i]) : n.filter((o) => !ds(o, i));
      } else
        return r.target.checked;
    else {
      if (e.tagName.toLowerCase() === "select" && e.multiple)
        return t.includes("number") ? Array.from(r.target.selectedOptions).map((i) => {
          let o = i.value || i.text;
          return xt(o);
        }) : t.includes("boolean") ? Array.from(r.target.selectedOptions).map((i) => {
          let o = i.value || i.text;
          return qe(o);
        }) : Array.from(r.target.selectedOptions).map((i) => i.value || i.text);
      {
        let i;
        return Wn(e) ? r.target.checked ? i = r.target.value : i = n : i = r.target.value, t.includes("number") ? xt(i) : t.includes("boolean") ? qe(i) : t.includes("trim") ? i.trim() : i;
      }
    }
  });
}
function xt(e) {
  let t = e ? parseFloat(e) : null;
  return ps(t) ? t : e;
}
function ds(e, t) {
  return e == t;
}
function ps(e) {
  return !Array.isArray(e) && !isNaN(e);
}
function Wr(e) {
  return e !== null && typeof e == "object" && typeof e.get == "function" && typeof e.set == "function";
}
O("cloak", (e) => queueMicrotask(() => k(() => e.removeAttribute(xe("cloak")))));
Rn(() => `[${xe("init")}]`);
O("init", re((e, { expression: t }, { evaluate: r }) => typeof t == "string" ? !!t.trim() && r(t, {}, !1) : r(t, {}, !1)));
O("text", (e, { expression: t }, { effect: r, evaluateLater: n }) => {
  let i = n(t);
  r(() => {
    i((o) => {
      k(() => {
        e.textContent = o;
      });
    });
  });
});
O("html", (e, { expression: t }, { effect: r, evaluateLater: n }) => {
  let i = n(t);
  r(() => {
    i((o) => {
      k(() => {
        e.innerHTML = o, e._x_ignoreSelf = !0, G(e), delete e._x_ignoreSelf;
      });
    });
  });
});
ir(_n(":", vn(xe("bind:"))));
var ui = (e, { value: t, modifiers: r, expression: n, original: i }, { effect: o, cleanup: a }) => {
  if (!t) {
    let l = {};
    ya(l), j(e, n)((f) => {
      qn(e, f, i);
    }, { scope: l });
    return;
  }
  if (t === "key")
    return hs(e, n);
  if (e._x_inlineBindings && e._x_inlineBindings[t] && e._x_inlineBindings[t].extract)
    return;
  let s = j(e, n);
  o(() => s((l) => {
    l === void 0 && typeof n == "string" && n.match(/\./) && (l = ""), k(() => Ln(e, t, l, r));
  })), a(() => {
    e._x_undoAddedClasses && e._x_undoAddedClasses(), e._x_undoAddedStyles && e._x_undoAddedStyles();
  });
};
ui.inline = (e, { value: t, modifiers: r, expression: n }) => {
  t && (e._x_inlineBindings || (e._x_inlineBindings = {}), e._x_inlineBindings[t] = { expression: n, extract: !1 });
};
O("bind", ui);
function hs(e, t) {
  e._x_keyExpression = t;
}
$n(() => `[${xe("data")}]`);
O("data", (e, { expression: t }, { cleanup: r }) => {
  if (gs(e))
    return;
  t = t === "" ? "{}" : t;
  let n = {};
  Mt(n, e);
  let i = {};
  wa(i, n);
  let o = ce(e, t, { scope: i });
  (o === void 0 || o === !0) && (o = {}), Mt(o, e);
  let a = be(o);
  dn(a);
  let s = Ie(e, a);
  a.init && ce(e, a.init), r(() => {
    a.destroy && ce(e, a.destroy), s();
  });
});
rt((e, t) => {
  e._x_dataStack && (t._x_dataStack = e._x_dataStack, t.setAttribute("data-has-alpine-state", !0));
});
function gs(e) {
  return ee ? Pt ? !0 : e.hasAttribute("data-has-alpine-state") : !1;
}
O("show", (e, { modifiers: t, expression: r }, { effect: n }) => {
  let i = j(e, r);
  e._x_doHide || (e._x_doHide = () => {
    k(() => {
      e.style.setProperty("display", "none", t.includes("important") ? "important" : void 0);
    });
  }), e._x_doShow || (e._x_doShow = () => {
    k(() => {
      e.style.length === 1 && e.style.display === "none" ? e.removeAttribute("style") : e.style.removeProperty("display");
    });
  });
  let o = () => {
    e._x_doHide(), e._x_isShown = !1;
  }, a = () => {
    e._x_doShow(), e._x_isShown = !0;
  }, s = () => setTimeout(a), l = Lt(
    (u) => u ? a() : o(),
    (u) => {
      typeof e._x_toggleAndCascadeWithTransitions == "function" ? e._x_toggleAndCascadeWithTransitions(e, u, a, o) : u ? s() : o();
    }
  ), c, f = !0;
  n(() => i((u) => {
    !f && u === c || (t.includes("immediate") && (u ? s() : o()), l(u), c = u, f = !1);
  }));
});
O("for", (e, { expression: t }, { effect: r, cleanup: n }) => {
  let i = bs(t), o = j(e, i.items), a = j(
    e,
    // the x-bind:key expression is stored for our use instead of evaluated.
    e._x_keyExpression || "index"
  );
  e._x_prevKeys = [], e._x_lookup = {}, r(() => ms(e, i, o, a)), n(() => {
    Object.values(e._x_lookup).forEach((s) => k(
      () => {
        _e(s), s.remove();
      }
    )), delete e._x_prevKeys, delete e._x_lookup;
  });
});
function ms(e, t, r, n) {
  let i = (a) => typeof a == "object" && !Array.isArray(a), o = e;
  r((a) => {
    ys(a) && a >= 0 && (a = Array.from(Array(a).keys(), (g) => g + 1)), a === void 0 && (a = []);
    let s = e._x_lookup, l = e._x_prevKeys, c = [], f = [];
    if (i(a))
      a = Object.entries(a).map(([g, w]) => {
        let m = Br(t, w, g, a);
        n((y) => {
          f.includes(y) && W("Duplicate key on x-for", e), f.push(y);
        }, { scope: { index: g, ...m } }), c.push(m);
      });
    else
      for (let g = 0; g < a.length; g++) {
        let w = Br(t, a[g], g, a);
        n((m) => {
          f.includes(m) && W("Duplicate key on x-for", e), f.push(m);
        }, { scope: { index: g, ...w } }), c.push(w);
      }
    let u = [], h = [], b = [], _ = [];
    for (let g = 0; g < l.length; g++) {
      let w = l[g];
      f.indexOf(w) === -1 && b.push(w);
    }
    l = l.filter((g) => !b.includes(g));
    let C = "template";
    for (let g = 0; g < f.length; g++) {
      let w = f[g], m = l.indexOf(w);
      if (m === -1)
        l.splice(g, 0, w), u.push([C, g]);
      else if (m !== g) {
        let y = l.splice(g, 1)[0], S = l.splice(m - 1, 1)[0];
        l.splice(g, 0, S), l.splice(m, 0, y), h.push([y, S]);
      } else
        _.push(w);
      C = w;
    }
    for (let g = 0; g < b.length; g++) {
      let w = b[g];
      w in s && (k(() => {
        _e(s[w]), s[w].remove();
      }), delete s[w]);
    }
    for (let g = 0; g < h.length; g++) {
      let [w, m] = h[g], y = s[w], S = s[m], q = document.createElement("div");
      k(() => {
        S || W('x-for ":key" is undefined or invalid', o, m, s), S.after(q), y.after(S), S._x_currentIfEl && S.after(S._x_currentIfEl), q.before(y), y._x_currentIfEl && y.after(y._x_currentIfEl), q.remove();
      }), S._x_refreshXForScope(c[f.indexOf(m)]);
    }
    for (let g = 0; g < u.length; g++) {
      let [w, m] = u[g], y = w === "template" ? o : s[w];
      y._x_currentIfEl && (y = y._x_currentIfEl);
      let S = c[m], q = f[m], P = document.importNode(o.content, !0).firstElementChild, F = be(S);
      Ie(P, F, o), P._x_refreshXForScope = (v) => {
        Object.entries(v).forEach(([T, I]) => {
          F[T] = I;
        });
      }, k(() => {
        y.after(P), re(() => G(P))();
      }), typeof q == "object" && W("x-for key cannot be an object, it must be a string or an integer", o), s[q] = P;
    }
    for (let g = 0; g < _.length; g++)
      s[_[g]]._x_refreshXForScope(c[f.indexOf(_[g])]);
    o._x_prevKeys = f;
  });
}
function bs(e) {
  let t = /,([^,\}\]]*)(?:,([^,\}\]]*))?$/, r = /^\s*\(|\)\s*$/g, n = /([\s\S]*?)\s+(?:in|of)\s+([\s\S]*)/, i = e.match(n);
  if (!i)
    return;
  let o = {};
  o.items = i[2].trim();
  let a = i[1].replace(r, "").trim(), s = a.match(t);
  return s ? (o.item = a.replace(t, "").trim(), o.index = s[1].trim(), s[2] && (o.collection = s[2].trim())) : o.item = a, o;
}
function Br(e, t, r, n) {
  let i = {};
  return /^\[.*\]$/.test(e.item) && Array.isArray(t) ? e.item.replace("[", "").replace("]", "").split(",").map((a) => a.trim()).forEach((a, s) => {
    i[a] = t[s];
  }) : /^\{.*\}$/.test(e.item) && !Array.isArray(t) && typeof t == "object" ? e.item.replace("{", "").replace("}", "").split(",").map((a) => a.trim()).forEach((a) => {
    i[a] = t[a];
  }) : i[e.item] = t, e.index && (i[e.index] = r), e.collection && (i[e.collection] = n), i;
}
function ys(e) {
  return !Array.isArray(e) && !isNaN(e);
}
function di() {
}
di.inline = (e, { expression: t }, { cleanup: r }) => {
  let n = et(e);
  n._x_refs || (n._x_refs = {}), n._x_refs[t] = e, r(() => delete n._x_refs[t]);
};
O("ref", di);
O("if", (e, { expression: t }, { effect: r, cleanup: n }) => {
  e.tagName.toLowerCase() !== "template" && W("x-if can only be used on a <template> tag", e);
  let i = j(e, t), o = () => {
    if (e._x_currentIfEl)
      return e._x_currentIfEl;
    let s = e.content.cloneNode(!0).firstElementChild;
    return Ie(s, {}, e), k(() => {
      e.after(s), re(() => G(s))();
    }), e._x_currentIfEl = s, e._x_undoIf = () => {
      k(() => {
        _e(s), s.remove();
      }), delete e._x_currentIfEl;
    }, s;
  }, a = () => {
    e._x_undoIf && (e._x_undoIf(), delete e._x_undoIf);
  };
  r(() => i((s) => {
    s ? o() : a();
  })), n(() => e._x_undoIf && e._x_undoIf());
});
O("id", (e, { expression: t }, { evaluate: r }) => {
  r(t).forEach((i) => is(e, i));
});
rt((e, t) => {
  e._x_ids && (t._x_ids = e._x_ids);
});
ir(_n("@", vn(xe("on:"))));
O("on", re((e, { value: t, modifiers: r, expression: n }, { cleanup: i }) => {
  let o = n ? j(e, n) : () => {
  };
  e.tagName.toLowerCase() === "template" && (e._x_forwardEvents || (e._x_forwardEvents = []), e._x_forwardEvents.includes(t) || e._x_forwardEvents.push(t));
  let a = Vt(e, t, r, (s) => {
    o(() => {
    }, { scope: { $event: s }, params: [s] });
  });
  i(() => a());
}));
at("Collapse", "collapse", "collapse");
at("Intersect", "intersect", "intersect");
at("Focus", "trap", "focus");
at("Mask", "mask", "mask");
function at(e, t, r) {
  O(t, (n) => W(`You can't use [x-${t}] without first installing the "${e}" plugin here: https://alpinejs.dev/plugins/${r}`, n));
}
Le.setEvaluator(bn);
Le.setReactivityEngine({ reactive: mr, effect: $a, release: Ra, raw: A });
var xs = Le, Vr = xs;
(function(e) {
  var a, s, l;
  const t = [], r = [];
  n(((a = e.TwindScope) == null ? void 0 : a.style) || [], ((s = e.TwindScope) == null ? void 0 : s.script) || []);
  function n(c, f) {
    c.forEach((u) => {
      let h = "inlineStyle";
      /^https?:\/\//.test(u) && (h = "url"), t.push({
        type: h,
        str: u
      });
    }), f.forEach((u) => {
      let h = "inlineScript";
      /^https?:\/\//.test(u) && (h = "url"), r.push({
        type: h,
        str: u
      });
    });
  }
  const i = Oi(
    Zr({
      presets: [Ii(), co()],
      ...((l = e.TwindScope) == null ? void 0 : l.config) || {}
    })
  );
  class o extends i(HTMLElement) {
    constructor() {
      super(), this.attachShadow({ mode: "open" }), this.integrateStyleAndScript(), this.shadowRoot && (this.shadowRoot.innerHTML = this.innerHTML, this.innerHTML = "", Vr.initTree(this.shadowRoot.firstElementChild));
    }
    disconnectedCallback() {
      this.shadowRoot && Vr.destroyTree(this.shadowRoot.firstElementChild);
    }
    integrateStyleAndScript() {
      t.length > 0 && t.forEach((f) => {
        if (this.shadowRoot) {
          if (f.type === "inlineStyle") {
            const u = new CSSStyleSheet();
            u.replaceSync(f.str), this.shadowRoot.adoptedStyleSheets = [u];
          }
          if (f.type === "url") {
            const u = document.createElement("link");
            u.rel = "stylesheet", u.href = f.str, this.shadowRoot.appendChild(u);
          }
        }
      }), r.length > 0 && r.forEach((f) => {
        if (this.shadowRoot) {
          if (f.type === "inlineScript") {
            const u = document.createElement("script");
            u.textContent = f.str, this.shadowRoot.appendChild(u);
          }
          if (f.type === "url") {
            const u = document.createElement("script");
            u.src = f.str, this.shadowRoot.appendChild(u);
          }
        }
      });
    }
  }
  customElements.define("twind-scope", o);
})(window);

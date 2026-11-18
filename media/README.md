
## Scoping the media library

`Searcher` is the isolation boundary. It is applied to every `media_libraries`
query in this package — the listing and chooser grid, folder trees, breadcrumbs,
folder item counts, by-id reads, and the model's generic presets CRUD/search
event funcs. Upload, new-folder and move targets are validated against it too.

```go
media.New(db).Searcher(func(db *gorm.DB, ctx *web.EventContext) *gorm.DB {
    return db.Where("tenant_id = ?", currentTenantID(ctx.R))
})
```

`CurrentUserID` is **not** an isolation boundary. It only supplies a
`user_id = ?` filter to the listing grid when no `Searcher` is configured, so
rows stay reachable by id. That is deliberate: some roles are meant to see every
user's uploads while the grid still defaults to their own. Configure a `Searcher`
whenever rows must actually be isolated.

Two things a `Searcher` does not cover:

- `MediaBoxSetterFunc` writes whatever `MediaBox` JSON the form submits, so a
  known URL can still be attached to a field without reading the row.
- Anything an application queries from `media_libraries` on its own.

## Cropping logic explanation

The principles of cropping are as follows

1. Users get what they see in the media library selector after selecting an image. No matter it is in the PageBuilder or in SEO or other places
2. Cropping an image won't affect any other places that use the same image

- Crop will always create a new file with rule `{file}_{size}_{crop_id}.ext`
    - original file is `file.png` after crop will be `file_{uuid}.png`
    - original file is `file_thumb.png` after crop will be `file_thumb_{uuid}.png`
- `cropOption` will not be saved in the media library table; `cropOption` will be saved where the image is used.
    - If the image is used in the SEO configuration. `cropOption` will be saved in the SEO record, the field name is `OpenGraphImageFromMediaLibrary`. it is a `MediaBox`
    - If the image is used in the PageBuilder. `cropOption` will be saved in the PageBuilder record, it is a `MediaBox`


## To cleanup unused file copies

The cropping logic might leave some unused file copies. If we need to clean them up. We have to fetch all MediaBox records and compare it with the file names in the file system. Then remove the unused files.

## Upload validation

Uploads through the media library are checked before they are stored:

- File names ending in a dot or whitespace, or containing control characters,
  are rejected.
- Files with script-capable extensions are rejected by default: the static
  list in `base.DeniedExtensions()` (`.svg`, `.svgz`, `.html`, `.htm`,
  `.xhtml`, `.xht`, `.xml`, `.xsl`, `.xslt`, `.js`, `.mjs`, `.php`, `.phtml`,
  `.shtml`, `.swf`, `.ehtml`, `.rdf`, `.xbl`, `.xsd`, `.rss`, `.atom`, `.xul`,
  `.mml`, `.mht`, `.mhtml`, `.wsdl`), plus any extension whose MIME type
  (`mime.TypeByExtension`) is HTML, XML, SVG or JavaScript.
- The leading bytes are sniffed: files classified as images must be a raster
  image and files classified as videos must not be text. Content sniffed as
  HTML/SVG/XML is rejected for every type, and plain files (or extensions
  without a known MIME type) must not start with `<`.
- `AllowedExtensions(...)` restricts uploads to the listed extensions
  (case-insensitive, with or without the dot). Listing a denied extension such
  as `svg` explicitly allows it and skips the markup check for it; do that only
  when the storage serves those files safely (separate domain, or
  `Content-Disposition: attachment`). Allowlisting any other extension never
  relaxes the markup check.

```go
media.New(db).AllowedExtensions("jpg", "jpeg", "png", "gif", "mp4", "pdf")
```

## Deleting stored files

Deleting media library items only deletes their rows by default, because
`MediaBox` values elsewhere keep their own copy of the file URL. Opt in to also
remove the stored files from `oss.Storage`:

```go
media.New(db).DeleteFilesOnDelete(true)
```

- When the URL template gives each row its own directory — a `{{uuid}}`
  segment, or `{{primary_key}}` right after `{{class}}` as in the default
  `/system/{{class}}/{{primary_key}}/...` — everything under that directory is
  removed, including crop copies.
- Otherwise only the file, its original copy and its size variants are
  removed; crop copies (`{file}_{crop_id}.ext`) are not tracked on the row and
  stay.

Storage errors are logged and ignored. Media library rows are soft-deleted;
with this option a restored row has no files.

## Unguessable storage URLs

Besides `{{class}}`, `{{primary_key}}`, `{{column}}`, `{{filename}}`,
`{{basename}}`, `{{extension}}`, `{{hash}}`, `{{short_hash}}`,
`{{filename_with_hash}}` and `{{filename_with_short_hash}}`, URL templates
support `{{uuid}}`: a random UUID v4, generated once per upload, so every size
variant of an upload shares it. Use it instead of `{{primary_key}}` to make
media URLs non-enumerable:

```go
// overrides the template of every media field saved by the media library
// saver, including the MediaLibrary.File tag template
base.MediaLibraryURL = "/system/{{class}}/{{uuid}}/{{column}}.{{extension}}"
// or only the fallback for media fields without a `mediaLibrary:"url:..."` tag
oss.URLTemplate = "/system/{{class}}/{{uuid}}/{{column}}/{{filename_with_hash}}"
```

Re-rendering the URL without a new upload (cropping) keeps the existing UUID,
so crop copies stay in the same directory; a new upload gets a new UUID.

Keep a per-row component (`{{uuid}}` or `{{primary_key}}`) in the path: a
template without one, e.g. `/system/{{class}}/{{column}}.{{extension}}`, makes
different rows write to — and overwrite — the same files.

The default templates are unchanged.

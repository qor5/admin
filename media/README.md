
## New Crop

- crop will create new file with rule `{file}_{size}_{crop_id}.ext`
    -  eg. `file_xx-xx.png`   `file.thumb.xx-xx.png`
- `media library` in db will don`t save croOption;use image field will save cropOption
- `media library` will save crop_ids to related use image;but delete use image won`t pop it
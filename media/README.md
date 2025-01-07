
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

local k8s = import 'k8s.jsonnet';

local ecr_prefix = 'public.ecr.aws/qor5/';

local images = [
  { type: 'deployment', name: 'example', image: ecr_prefix + 'example' },
  { type: 'deployment', name: 'publisher', image: ecr_prefix + 'publisher' },
];

k8s.set_images(
  namespace='qor5-test',
  images=images
)

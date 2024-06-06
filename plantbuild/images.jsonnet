local k8s = import 'k8s.jsonnet';

local ecr_prefix = 'public.ecr.aws/qor5/';

local images = [
  { type: 'deployment', name: 'docs', image: ecr_prefix + 'docs' },
  { type: 'deployment', name: 'example', image: ecr_prefix + 'example' },
  { type: 'deployment', name: 'publisher', image: ecr_prefix + 'publisher' },
  { type: 'deployment', name: 'data-resetor', image: ecr_prefix + 'data-resetor' },
];

k8s.set_images(
  namespace='qor5-test',
  images=images
)

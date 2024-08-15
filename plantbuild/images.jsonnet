local k8s = import 'k8s.jsonnet';

local ecr_prefix = 'public.ecr.aws/qor5/';

local images = [
  { type: 'deployment', name: 'docs', image: ecr_prefix + 'docs' },
];

k8s.set_images(
  namespace='qor5-test',
  images=images
)

local c = import 'dc.jsonnet';
local dc = c {
  dockerRegistry: 'public.ecr.aws',
};

dc.build_apps_image('theplant/qor5', [
  { name: 'docs', dockerfile: './docs/Dockerfile', context: '.' },
  { name: 'example', dockerfile: './example/Dockerfile', context: '.' },
  { name: 'publisher', dockerfile: './example/cmd/publisher/Dockerfile', context: '.' },
  { name: 'data-resetor', dockerfile: './example/cmd/data-resetor/Dockerfile', context: '.' },
])

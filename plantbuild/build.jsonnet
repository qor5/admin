local c = import 'dc.jsonnet';
local dc = c {
  dockerRegistry: 'public.ecr.aws',
};

dc.build_apps_image('theplant/qor5', [
  { name: 'example', dockerfile: './example/Dockerfile', context: '.' },
  { name: 'publisher', dockerfile: './example/cmd/publisher/Dockerfile', context: '.' },
])

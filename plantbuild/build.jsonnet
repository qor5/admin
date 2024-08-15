local c = import 'dc.jsonnet';
local dc = c {
  dockerRegistry: 'public.ecr.aws',
};

dc.build_apps_image('theplant/qor5', [
  { name: 'docs', dockerfile: './docs/Dockerfile', context: '.' },
])

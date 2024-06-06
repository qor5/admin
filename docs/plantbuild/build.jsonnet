local c = import 'dc.jsonnet';
local dc = c {
  dockerRegistry: 'public.ecr.aws',
};

dc.build_apps_image('qor5/qor5', [
  { name: 'docs', dockerfile: './Dockerfile', context: '.' },
])

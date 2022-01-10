local c = import 'dc.jsonnet';
local dc = c {
  dockerRegistry: '562055475000.dkr.ecr.ap-northeast-1.amazonaws.com',
};

dc.build_apps_image('theplant/qor5', [
  { name: 'example', dockerfile: './example/Dockerfile', context: './' },
  { name: 'publisher', dockerfile: './example/cmd/publisher/Dockerfile', context: './' },
])

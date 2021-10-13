Vue.component('v-seo', {
    template: "<div><slot></slot></div>",
    data: function () {
        return {
            vnode: null,   
        }
      },
    methods: {
        tagInputsFocus: function(v){
            vnode = v
        },
        addTags: function(tag){    
            var lazyValue = vnode.$data.lazyValue
            var startString = lazyValue.substring(0, vnode.$refs.input.selectionStart);
            var endString = lazyValue.substring(vnode.$refs.input.selectionEnd, lazyValue.length);
            
            vnode.$data.lazyValue = startString + '{{' + tag +'}}' + endString;
            vnode.$emit('input', vnode.$data.lazyValue); 
            vnode.focus();
        },
    },
});

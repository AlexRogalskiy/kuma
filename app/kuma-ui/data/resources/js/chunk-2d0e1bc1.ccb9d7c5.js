(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-2d0e1bc1"],{"7c65":function(t,e,n){"use strict";n.r(e);var i=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"dataplanes-detail"},[n("YamlView",{attrs:{title:"Entity Overview",content:t.entity}})],1)},a=[],o=n("ff9d"),r=n("be10"),s={name:"TrafficRouteDetail",metaInfo:{title:"Traffic Route Details"},components:{MetricGrid:r["a"],YamlView:o["a"]},data:function(){return{entity:null}},watch:{$route:function(t,e){this.bootstrap()}},beforeMount:function(){this.bootstrap()},methods:{bootstrap:function(){var t=this,e=this.$route.params.mesh,n=this.$route.params.trafficroute;return this.$api.getTrafficRoute(e,n).then((function(e){e?t.entity=e:t.$router.push("/404")})).catch((function(e){console.error(e),t.entity=e}))}}},c=s,u=n("2877"),f=Object(u["a"])(c,i,a,!1,null,null,null);e["default"]=f.exports}}]);
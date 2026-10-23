const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["assets/ProfilesView-DTwSR_35.js","assets/radix-C072bb2c.js","assets/lucide-DYK3be8Z.js","assets/card-DdC5TnK2.js","assets/tabs-CATgr_fF.js","assets/segmented-control-CfywDhUz.js","assets/select-dropdown-CqgDoGFb.js","assets/AgentBadge-CJsk9-Mf.js","assets/ResumeModal-DrQK4oa2.js","assets/SessionsView-BJNgU4jK.js","assets/table-o9qkV6Iu.js","assets/SessionDetailView-CqdZp2XU.js","assets/markdown-WYmnjSVf.js","assets/McpView-vcf5UOVw.js","assets/NetworkView-4sBoSKvp.js"])))=>i.map(i=>d[i]);
var eb=Object.defineProperty;var tb=(s,r,c)=>r in s?eb(s,r,{enumerable:!0,configurable:!0,writable:!0,value:c}):s[r]=c;var y0=(s,r,c)=>tb(s,typeof r!="symbol"?r+"":r,c);import{r as nb,j as o,D as hp,a as lb,b as gp,c as ob,d as pp,e as yp,f as sb,S as ab,V as xp,R as bp,A as vp,C as wp,T as Sp,g as jp,P as rb,h as L_}from"./radix-C072bb2c.js";import{r as ib,g as cb,a as h,S as ub,L as db,T as kp,b as fb,c as x0,P as _b,X as Cp,A as b0,C as mb,F as hb,d as i_,e as c_,f as gb,h as $r,i as pb,j as yb,I as xb,k as Np,l as bb,m as vb,n as wb,o as Sb,R as jb,p as Ji,q as kb}from"./lucide-DYK3be8Z.js";(function(){const r=document.createElement("link").relList;if(r&&r.supports&&r.supports("modulepreload"))return;for(const f of document.querySelectorAll('link[rel="modulepreload"]'))u(f);new MutationObserver(f=>{for(const d of f)if(d.type==="childList")for(const m of d.addedNodes)m.tagName==="LINK"&&m.rel==="modulepreload"&&u(m)}).observe(document,{childList:!0,subtree:!0});function c(f){const d={};return f.integrity&&(d.integrity=f.integrity),f.referrerPolicy&&(d.referrerPolicy=f.referrerPolicy),f.crossOrigin==="use-credentials"?d.credentials="include":f.crossOrigin==="anonymous"?d.credentials="omit":d.credentials="same-origin",d}function u(f){if(f.ep)return;f.ep=!0;const d=c(f);fetch(f.href,d)}})();var u_={exports:{}},Pi={},d_={exports:{}},f_={};/**
 * @license React
 * scheduler.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var v0;function Cb(){return v0||(v0=1,(function(s){function r(G,we){var fe=G.length;G.push(we);e:for(;0<fe;){var Oe=fe-1>>>1,Se=G[Oe];if(0<f(Se,we))G[Oe]=we,G[fe]=Se,fe=Oe;else break e}}function c(G){return G.length===0?null:G[0]}function u(G){if(G.length===0)return null;var we=G[0],fe=G.pop();if(fe!==we){G[0]=fe;e:for(var Oe=0,Se=G.length,Qe=Se>>>1;Oe<Qe;){var Ut=2*(Oe+1)-1,Ye=G[Ut],Ne=Ut+1,Xt=G[Ne];if(0>f(Ye,fe))Ne<Se&&0>f(Xt,Ye)?(G[Oe]=Xt,G[Ne]=fe,Oe=Ne):(G[Oe]=Ye,G[Ut]=fe,Oe=Ut);else if(Ne<Se&&0>f(Xt,fe))G[Oe]=Xt,G[Ne]=fe,Oe=Ne;else break e}}return we}function f(G,we){var fe=G.sortIndex-we.sortIndex;return fe!==0?fe:G.id-we.id}if(s.unstable_now=void 0,typeof performance=="object"&&typeof performance.now=="function"){var d=performance;s.unstable_now=function(){return d.now()}}else{var m=Date,g=m.now();s.unstable_now=function(){return m.now()-g}}var p=[],j=[],k=1,b=null,v=3,z=!1,E=!1,$=!1,B=!1,F=typeof setTimeout=="function"?setTimeout:null,Me=typeof clearTimeout=="function"?clearTimeout:null,te=typeof setImmediate<"u"?setImmediate:null;function C(G){for(var we=c(j);we!==null;){if(we.callback===null)u(j);else if(we.startTime<=G)u(j),we.sortIndex=we.expirationTime,r(p,we);else break;we=c(j)}}function ie(G){if($=!1,C(G),!E)if(c(p)!==null)E=!0,de||(de=!0,V());else{var we=c(j);we!==null&&ue(ie,we.startTime-G)}}var de=!1,W=-1,Ee=5,ne=-1;function Ce(){return B?!0:!(s.unstable_now()-ne<Ee)}function pe(){if(B=!1,de){var G=s.unstable_now();ne=G;var we=!0;try{e:{E=!1,$&&($=!1,Me(W),W=-1),z=!0;var fe=v;try{t:{for(C(G),b=c(p);b!==null&&!(b.expirationTime>G&&Ce());){var Oe=b.callback;if(typeof Oe=="function"){b.callback=null,v=b.priorityLevel;var Se=Oe(b.expirationTime<=G);if(G=s.unstable_now(),typeof Se=="function"){b.callback=Se,C(G),we=!0;break t}b===c(p)&&u(p),C(G)}else u(p);b=c(p)}if(b!==null)we=!0;else{var Qe=c(j);Qe!==null&&ue(ie,Qe.startTime-G),we=!1}}break e}finally{b=null,v=fe,z=!1}we=void 0}}finally{we?V():de=!1}}}var V;if(typeof te=="function")V=function(){te(pe)};else if(typeof MessageChannel<"u"){var Ie=new MessageChannel,he=Ie.port2;Ie.port1.onmessage=pe,V=function(){he.postMessage(null)}}else V=function(){F(pe,0)};function ue(G,we){W=F(function(){G(s.unstable_now())},we)}s.unstable_IdlePriority=5,s.unstable_ImmediatePriority=1,s.unstable_LowPriority=4,s.unstable_NormalPriority=3,s.unstable_Profiling=null,s.unstable_UserBlockingPriority=2,s.unstable_cancelCallback=function(G){G.callback=null},s.unstable_forceFrameRate=function(G){0>G||125<G?console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"):Ee=0<G?Math.floor(1e3/G):5},s.unstable_getCurrentPriorityLevel=function(){return v},s.unstable_next=function(G){switch(v){case 1:case 2:case 3:var we=3;break;default:we=v}var fe=v;v=we;try{return G()}finally{v=fe}},s.unstable_requestPaint=function(){B=!0},s.unstable_runWithPriority=function(G,we){switch(G){case 1:case 2:case 3:case 4:case 5:break;default:G=3}var fe=v;v=G;try{return we()}finally{v=fe}},s.unstable_scheduleCallback=function(G,we,fe){var Oe=s.unstable_now();switch(typeof fe=="object"&&fe!==null?(fe=fe.delay,fe=typeof fe=="number"&&0<fe?Oe+fe:Oe):fe=Oe,G){case 1:var Se=-1;break;case 2:Se=250;break;case 5:Se=1073741823;break;case 4:Se=1e4;break;default:Se=5e3}return Se=fe+Se,G={id:k++,callback:we,priorityLevel:G,startTime:fe,expirationTime:Se,sortIndex:-1},fe>Oe?(G.sortIndex=fe,r(j,G),c(p)===null&&G===c(j)&&($?(Me(W),W=-1):$=!0,ue(ie,fe-Oe))):(G.sortIndex=Se,r(p,G),E||z||(E=!0,de||(de=!0,V()))),G},s.unstable_shouldYield=Ce,s.unstable_wrapCallback=function(G){var we=v;return function(){var fe=v;v=we;try{return G.apply(this,arguments)}finally{v=fe}}}})(f_)),f_}var w0;function Nb(){return w0||(w0=1,d_.exports=Cb()),d_.exports}/**
 * @license React
 * react-dom-client.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var S0;function Mb(){if(S0)return Pi;S0=1;var s=Nb(),r=ib(),c=nb();function u(e){var t="https://react.dev/errors/"+e;if(1<arguments.length){t+="?args[]="+encodeURIComponent(arguments[1]);for(var n=2;n<arguments.length;n++)t+="&args[]="+encodeURIComponent(arguments[n])}return"Minified React error #"+e+"; visit "+t+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}function f(e){return!(!e||e.nodeType!==1&&e.nodeType!==9&&e.nodeType!==11)}function d(e){for(var t=e,n=t;n&&!n.alternate;)t=n,(t.flags&4098)!==0&&(e=t.return),n=t.return;for(;t.return;)t=t.return;return t.tag===3?e:null}function m(e){if(e.tag===13){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function g(e){if(e.tag===31){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function p(e){if(d(e)!==e)throw Error(u(188))}function j(e){var t=e.alternate;if(!t){if(t=d(e),t===null)throw Error(u(188));return t!==e?null:e}for(var n=e,l=t;;){var a=n.return;if(a===null)break;var i=a.alternate;if(i===null){if(l=a.return,l!==null){n=l;continue}break}if(a.child===i.child){for(i=a.child;i;){if(i===n)return p(a),e;if(i===l)return p(a),t;i=i.sibling}throw Error(u(188))}if(n.return!==l.return)n=a,l=i;else{for(var _=!1,y=a.child;y;){if(y===n){_=!0,n=a,l=i;break}if(y===l){_=!0,l=a,n=i;break}y=y.sibling}if(!_){for(y=i.child;y;){if(y===n){_=!0,n=i,l=a;break}if(y===l){_=!0,l=i,n=a;break}y=y.sibling}if(!_)throw Error(u(189))}}if(n.alternate!==l)throw Error(u(190))}if(n.tag!==3)throw Error(u(188));return n.stateNode.current===n?e:t}function k(e){var t=e.tag;if(t===5||t===26||t===27||t===6)return e;for(e=e.child;e!==null;){if(t=k(e),t!==null)return t;e=e.sibling}return null}function b(e,t,n,l,a,i){for(;e!==null;){if((e.tag===5||e.tag===27||e.tag===6)&&n(e,l,a,i)||(e.tag!==22||e.memoizedState===null)&&(t||e.tag!==5&&e.tag!==27)&&b(e.child,t,n,l,a,i))return!0;e=e.sibling}return!1}function v(e){for(e=e.return;e!==null;){if(e.tag===3||e.tag===5||e.tag===27)return e;e=e.return}return null}function z(e){var t=!1;for(e=e.return;e!==null&&(e.tag===4&&(t=!0),!(e.tag===3||e.tag===5||e.tag===27));)e=e.return;return t}function E(e){var t=[null,null],n=v(e);return n===null||$(t,e,n.child,{foundSelf:!1}),t}function $(e,t,n,l){for(;n!==null;){if(n===t)l.foundSelf=!0;else if(n.tag===5||n.tag===27||n.tag===6){if(l.foundSelf)return e[1]=n,!0;e[0]=n}else if((n.tag!==22||n.memoizedState===null)&&$(e,t,n.child,l))return!0;n=n.sibling}return!1}function B(e){switch(e.tag){case 5:case 27:case 6:return e.stateNode;case 3:return e.stateNode.containerInfo;default:throw Error(u(559))}}var F=null,Me=null;function te(e,t,n){return e===n?!0:e===t?(F=e,!0):!1}function C(e,t,n){return e===n?(Me=e,!1):e===t?(Me!==null&&(F=e),!0):!1}function ie(e){if(e===null)return null;do e=e===null?null:e.return;while(e&&e.tag!==5&&e.tag!==27&&e.tag!==3);return e||null}function de(e,t,n){for(var l=0,a=e;a;a=n(a))l++;a=0;for(var i=t;i;i=n(i))a++;for(;0<l-a;)e=n(e),l--;for(;0<a-l;)t=n(t),a--;for(;l--;){if(e===t||t!==null&&e===t.alternate)return e;e=n(e),t=n(t)}return null}var W=Object.assign,Ee=Symbol.for("react.element"),ne=Symbol.for("react.transitional.element"),Ce=Symbol.for("react.portal"),pe=Symbol.for("react.fragment"),V=Symbol.for("react.strict_mode"),Ie=Symbol.for("react.profiler"),he=Symbol.for("react.consumer"),ue=Symbol.for("react.context"),G=Symbol.for("react.forward_ref"),we=Symbol.for("react.suspense"),fe=Symbol.for("react.suspense_list"),Oe=Symbol.for("react.memo"),Se=Symbol.for("react.lazy"),Qe=Symbol.for("react.activity"),Ut=Symbol.for("react.legacy_hidden"),Ye=Symbol.for("react.memo_cache_sentinel"),Ne=Symbol.for("react.view_transition"),Xt=Symbol.for("react.recoverable"),xe=Symbol.iterator;function We(e){return e===null||typeof e!="object"?null:(e=xe&&e[xe]||e["@@iterator"],typeof e=="function"?e:null)}var gn=Symbol.for("react.client.reference");function Qt(e){if(e==null)return null;if(typeof e=="function")return e.$$typeof===gn?null:e.displayName||e.name||null;if(typeof e=="string")return e;switch(e){case pe:return"Fragment";case Ie:return"Profiler";case V:return"StrictMode";case we:return"Suspense";case fe:return"SuspenseList";case Qe:return"Activity";case Ne:return"ViewTransition"}if(typeof e=="object")switch(e.$$typeof){case Ce:return"Portal";case ue:return e.displayName||"Context";case he:return(e._context.displayName||"Context")+".Consumer";case G:var t=e.render;return e=e.displayName,e||(e=t.displayName||t.name||"",e=e!==""?"ForwardRef("+e+")":"ForwardRef"),e;case Oe:return t=e.displayName||null,t!==null?t:Qt(e.type)||"Memo";case Se:t=e._payload,e=e._init;try{return Qt(e(t))}catch{}}return null}var mn=Array.isArray,Re=r.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,Ue=c.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,un={pending:!1,data:null,method:null,action:null},ul=[],Cn=-1;function bn(e){return{current:e}}function re(e){0>Cn||(e.current=ul[Cn],ul[Cn]=null,Cn--)}function ye(e,t){Cn++,ul[Cn]=e.current,e.current=t}var Ge=bn(null),nt=bn(null),lt=bn(null),rt=bn(null);function it(e,t){switch(ye(lt,t),ye(nt,e),ye(Ge,null),t.nodeType){case 9:case 11:e=(e=t.documentElement)&&(e=e.namespaceURI)?Ng(e):0;break;default:if(e=t.tagName,t=t.namespaceURI)t=Ng(t),e=Mg(t,e);else switch(e){case"svg":e=1;break;case"math":e=2;break;default:e=0}}re(Ge),ye(Ge,e)}function Fe(){re(Ge),re(nt),re(lt)}function Nt(e){var t=e.memoizedState;t!==null&&(Dr._currentValue=t.memoizedState,ye(rt,e)),t=Ge.current;var n=Mg(t,e.type);t!==n&&(ye(nt,e),ye(Ge,n))}function an(e){nt.current===e&&(re(Ge),re(nt)),rt.current===e&&(re(rt),Dr._currentValue=un)}var oe,O;function I(e){if(oe===void 0)try{throw Error()}catch(n){var t=n.stack.trim().match(/\n( *(at )?)/);oe=t&&t[1]||"",O=-1<n.stack.indexOf(`
    at`)?" (<anonymous>)":-1<n.stack.indexOf("@")?"@unknown:0:0":""}return`
`+oe+e+O}var P=!1;function ae(e,t){if(!e||P)return"";P=!0;var n=Error.prepareStackTrace;Error.prepareStackTrace=void 0;try{var l={DetermineComponentFrameRoot:function(){try{if(t){var J=function(){throw Error()};if(Object.defineProperty(J.prototype,"props",{set:function(){throw Error()}}),typeof Reflect=="object"&&Reflect.construct){try{Reflect.construct(J,[])}catch(ve){var A=ve}Reflect.construct(e,[],J)}else{try{J.call()}catch(ve){A=ve}J=!1;try{var q=Object.getOwnPropertyDescriptor(e.prototype,"props");Object.defineProperty(e.prototype,"props",{configurable:!0,set:function(){throw Error()}}),J=!0,new e}finally{J&&(q!==void 0?Object.defineProperty(e.prototype,"props",q):delete e.prototype.props)}}}else{try{throw Error()}catch(ve){A=ve}(J=e())&&typeof J.catch=="function"&&J.catch(function(){})}}catch(ve){if(ve&&A&&typeof ve.stack=="string")return[ve.stack,A.stack]}return[null,null]}};l.DetermineComponentFrameRoot.displayName="DetermineComponentFrameRoot";var a=Object.getOwnPropertyDescriptor(l.DetermineComponentFrameRoot,"name");a&&a.configurable&&Object.defineProperty(l.DetermineComponentFrameRoot,"name",{value:"DetermineComponentFrameRoot"});var i=l.DetermineComponentFrameRoot(),_=i[0],y=i[1];if(_&&y){var M=_.split(`
`),L=y.split(`
`);for(a=l=0;l<M.length&&!M[l].includes("DetermineComponentFrameRoot");)l++;for(;a<L.length&&!L[a].includes("DetermineComponentFrameRoot");)a++;if(l===M.length||a===L.length)for(l=M.length-1,a=L.length-1;1<=l&&0<=a&&M[l]!==L[a];)a--;for(;1<=l&&0<=a;l--,a--)if(M[l]!==L[a]){if(l!==1||a!==1)do if(l--,a--,0>a||M[l]!==L[a]){var X=`
`+M[l].replace(" at new "," at ");return e.displayName&&X.includes("<anonymous>")&&(X=X.replace("<anonymous>",e.displayName)),X}while(1<=l&&0<=a);break}}}finally{P=!1,Error.prepareStackTrace=n}return(n=e?e.displayName||e.name:"")?I(n):""}function Be(e,t){switch(e.tag){case 26:case 27:case 5:return I(e.type);case 16:return I("Lazy");case 13:return e.child!==t&&t!==null?I("Suspense Fallback"):I("Suspense");case 19:return I("SuspenseList");case 0:case 15:return ae(e.type,!1);case 11:return ae(e.type.render,!1);case 1:return ae(e.type,!0);case 31:return I("Activity");case 30:return I("ViewTransition");default:return""}}function ke(e){try{var t="",n=null;do t+=Be(e,n),n=e,e=e.return;while(e);return t}catch(l){return`
Error generating stack: `+l.message+`
`+l.stack}}var _e=Object.prototype.hasOwnProperty,at=s.unstable_scheduleCallback,ot=s.unstable_cancelCallback,ct=s.unstable_shouldYield,Ke=s.unstable_requestPaint,Ze=s.unstable_now,bt=s.unstable_getCurrentPriorityLevel,Bt=s.unstable_ImmediatePriority,pn=s.unstable_UserBlockingPriority,pt=s.unstable_NormalPriority,At=s.unstable_LowPriority,kt=s.unstable_IdlePriority,ht=s.log,Nn=s.unstable_setDisableYieldValue,vt=null,Je=null;function je(e){if(typeof ht=="function"&&Nn(e),Je&&typeof Je.setStrictMode=="function")try{Je.setStrictMode(vt,e)}catch{}}var Gt=Math.clz32?Math.clz32:Ql,vn=Math.log,ho=Math.LN2;function Ql(e){return e>>>=0,e===0?32:31-(vn(e)/ho|0)|0}var yn=256,Fn=262144,Hn=4194304;function Mn(e){var t=e&42;if(t!==0)return t;switch(e&-e){case 1:return 1;case 2:return 2;case 4:return 4;case 8:return 8;case 16:return 16;case 32:return 32;case 64:return 64;case 128:return 128;case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:return e&-e;case 262144:case 524288:case 1048576:case 2097152:return e&3932160;case 4194304:case 8388608:case 16777216:case 33554432:return e&62914560;case 67108864:return 67108864;case 134217728:return 134217728;case 268435456:return 268435456;case 536870912:return 536870912;case 1073741824:return 0;default:return e}}function dl(e,t,n){var l=e.pendingLanes;if(l===0)return 0;var a=0,i=e.suspendedLanes,_=e.pingedLanes;e=e.warmLanes;var y=l&134217727;return y!==0?(l=y&~i,l!==0?a=Mn(l):(_&=y,_!==0?a=Mn(_):n||(n=y&~e,n!==0&&(a=Mn(n))))):(y=l&~i,y!==0?a=Mn(y):_!==0?a=Mn(_):n||(n=l&~e,n!==0&&(a=Mn(n)))),a===0?0:t!==0&&t!==a&&(t&i)===0&&(i=a&-a,n=t&-t,i>=n||i===32&&(n&4194048)!==0)?t:a}function sl(e,t){return(e.pendingLanes&~(e.suspendedLanes&~e.pingedLanes)&t)===0}function Gr(e,t){(t&8)!==0&&(t|=t&32);var n=e.entangledLanes;if(n!==0)for(e=e.entanglements,n&=t;0<n;){var l=31-Gt(n),a=1<<l;t|=e[l],n&=~a}return t}function hc(e,t){switch(e){case 1:case 2:case 4:case 8:case 64:return t+250;case 16:case 32:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return t+5e3;case 4194304:case 8388608:case 16777216:case 33554432:return-1;case 67108864:case 134217728:case 268435456:case 536870912:case 1073741824:return-1;default:return-1}}function Vr(){var e=Hn;return Hn<<=1,(Hn&62914560)===0&&(Hn=4194304),e}function Ot(e){for(var t=[],n=0;31>n;n++)t.push(e);return t}function Ps(e,t){e.pendingLanes|=t,t!==268435456&&(e.suspendedLanes=0,e.pingedLanes=0,e.warmLanes=0)}function vl(e,t,n,l,a,i){var _=e.pendingLanes;e.pendingLanes=n,e.suspendedLanes=0,e.pingedLanes=0,e.warmLanes=0,e.expiredLanes&=n,e.entangledLanes&=n,e.errorRecoveryDisabledLanes&=n,e.shellSuspendCounter=0;var y=e.entanglements,M=e.expirationTimes,L=e.hiddenUpdates;for(n=_&~n;0<n;){var X=31-Gt(n),J=1<<X;y[X]=0,M[X]=-1;var A=L[X];if(A!==null)for(L[X]=null,X=0;X<A.length;X++){var q=A[X];q!==null&&(q.lane&=-536870913)}n&=~J}l!==0&&Fr(e,l,0),i!==0&&a===0&&e.tag!==0&&(e.suspendedLanes|=i&~(_&~t))}function Fr(e,t,n){e.pendingLanes|=t,e.suspendedLanes&=~t;var l=31-Gt(t);e.entangledLanes|=t,e.entanglements[l]=e.entanglements[l]|1073741824|n&261930}function Yn(e,t){var n=e.entangledLanes|=t;for(e=e.entanglements;n;){var l=31-Gt(n),a=1<<l;a&t|e[l]&t&&(e[l]|=t),n&=~a}}function go(e,t){var n=t&-t;return n=(n&42)!==0?1:Wr(n),(n&(e.suspendedLanes|t))!==0?0:n}function Wr(e){switch(e){case 2:e=1;break;case 8:e=4;break;case 32:e=16;break;case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:case 4194304:case 8388608:case 16777216:case 33554432:e=128;break;case 268435456:e=134217728;break;default:e=0}return e}function $a(e){return e&=-e,2<e?8<e?(e&134217727)!==0?32:268435456:8:2}function Kr(){var e=Ue.p;return e!==0?e:(e=window.event,e===void 0?32:d0(e.type))}function Ua(e,t){var n=Ue.p;try{return Ue.p=e,t()}finally{Ue.p=n}}var ut=Math.random().toString(36).slice(2),En="__reactFiber$"+ut,wn="__reactProps$"+ut,Gl="__reactContainer$"+ut,ft="__reactEvents$"+ut,eo="__reactListeners$"+ut,is="__reactHandles$"+ut,po="__reactResources$"+ut,Uo="__reactMarker$"+ut,Yt="__reactLoad$"+ut;function ea(e){delete e[En],delete e[wn],delete e[eo],delete e[is]}function yo(e){var t;if(t=e[En])return t;for(var n=e.parentNode;n;){if(t=n[Gl]||n[En]){if(n=t.alternate,t.child!==null||n!==null&&n.child!==null)for(e=Qg(e);e!==null;){if(n=e[En])return n;e=Qg(e)}return t}e=n,n=e.parentNode}return null}function xo(e){if(e=e[En]||e[Gl]){var t=e.tag;if(t===5||t===6||t===13||t===31||t===26||t===27||t===3)return e}return null}function cs(e){var t=e.tag;if(t===5||t===26||t===27||t===6)return e.stateNode;throw Error(u(33))}function us(e){var t=e[po];return t||(t=e[po]={hoistableStyles:new Map,hoistableScripts:new Map}),t}function Tn(e){e[Uo]=!0}function ad(e){e[Yt]=void 0}var qn=new Set,ta={};function Ho(e,t){ds(e,t),ds(e+"Capture",t)}function ds(e,t){for(ta[e]=t,e=0;e<t.length;e++)qn.add(t[e])}var be=RegExp("^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"),Rl={},Ha={};function Zr(e){return _e.call(Ha,e)?!0:_e.call(Rl,e)?!1:be.test(e)?Ha[e]=!0:(Rl[e]=!0,!1)}var yt=!1;function gc(){var e=yt;return yt=!1,e}function na(e,t,n){if(Zr(t))if(n===null)e.removeAttribute(t);else{switch(typeof n){case"undefined":case"function":case"symbol":e.removeAttribute(t);return;case"boolean":var l=t.toLowerCase().slice(0,5);if(l!=="data-"&&l!=="aria-"){e.removeAttribute(t);return}}e.setAttribute(t,n)}}function Ya(e,t,n){if(n===null)e.removeAttribute(t);else{switch(typeof n){case"undefined":case"function":case"symbol":case"boolean":e.removeAttribute(t);return}e.setAttribute(t,n)}}function to(e,t,n,l){if(l===null)e.removeAttribute(n);else{switch(typeof l){case"undefined":case"function":case"symbol":case"boolean":e.removeAttribute(n);return}e.setAttributeNS(t,n,l)}}function On(e){switch(typeof e){case"bigint":case"boolean":case"number":case"string":case"undefined":return e;case"object":return e;default:return""}}function Jr(e){var t=e.type;return(e=e.nodeName)&&e.toLowerCase()==="input"&&(t==="checkbox"||t==="radio")}function fs(e,t,n){var l=Object.getOwnPropertyDescriptor(e.constructor.prototype,t);if(!e.hasOwnProperty(t)&&typeof l<"u"&&typeof l.get=="function"&&typeof l.set=="function"){var a=l.get,i=l.set;return Object.defineProperty(e,t,{configurable:!0,get:function(){return a.call(this)},set:function(_){n=""+_,i.call(this,_)}}),Object.defineProperty(e,t,{enumerable:l.enumerable}),{getValue:function(){return n},setValue:function(_){n=""+_},stopTracking:function(){e._valueTracker=null,delete e[t]}}}}function qa(e){if(!e._valueTracker){var t=Jr(e)?"checked":"value";e._valueTracker=fs(e,t,""+e[t])}}function Pr(e){if(!e)return!1;var t=e._valueTracker;if(!t)return!0;var n=t.getValue(),l="";return e&&(l=Jr(e)?e.checked?"true":"false":e.value),e=l,e!==n?(t.setValue(e),!0):!1}var _s=/[\n"\\]/g;function Xn(e){return e.replace(_s,function(t){return"\\"+t.charCodeAt(0).toString(16)+" "})}function ms(e,t,n,l,a,i,_,y){e.name="",_!=null&&typeof _!="function"&&typeof _!="symbol"&&typeof _!="boolean"?e.type=_:e.removeAttribute("type"),t!=null?_==="number"?(t===0&&e.value===""||e.value!=t)&&(e.value=""+On(t)):e.value!==""+On(t)&&(e.value=""+On(t)):_!=="submit"&&_!=="reset"||e.removeAttribute("value"),t!=null?_==="number"&&e.value==t?bo(e,On(e.value)):bo(e,On(t)):n!=null?bo(e,On(n)):l!=null&&e.removeAttribute("value"),a==null&&i!=null&&(e.defaultChecked=!!i),a!=null&&(e.checked=a&&typeof a!="function"&&typeof a!="symbol"),y!=null&&typeof y!="function"&&typeof y!="symbol"&&typeof y!="boolean"?e.name=""+On(y):e.removeAttribute("name")}function wl(e,t,n,l,a,i,_,y){if(i!=null&&typeof i!="function"&&typeof i!="symbol"&&typeof i!="boolean"&&(e.type=i),t!=null||n!=null){if(!(i!=="submit"&&i!=="reset"||t!=null)){qa(e);return}n=n!=null?""+On(n):"",t=t!=null?""+On(t):n,y||t===e.value||(e.value=t),e.defaultValue=t}l=l??a,l=typeof l!="function"&&typeof l!="symbol"&&!!l,e.checked=y?e.checked:!!l,e.defaultChecked=!!l,_!=null&&typeof _!="function"&&typeof _!="symbol"&&typeof _!="boolean"&&(e.name=_),qa(e)}function bo(e,t){e.defaultValue!==""+t&&(e.defaultValue=""+t)}function Yo(e,t,n,l){if(e=e.options,t){t={};for(var a=0;a<n.length;a++)t["$"+n[a]]=!0;for(n=0;n<e.length;n++)a=t.hasOwnProperty("$"+e[n].value),e[n].selected!==a&&(e[n].selected=a),a&&l&&(e[n].defaultSelected=!0)}else{for(n=""+On(n),t=null,a=0;a<e.length;a++){if(e[a].value===n){e[a].selected=!0,l&&(e[a].defaultSelected=!0);return}t!==null||e[a].disabled||(t=e[a])}t!==null&&(t.selected=!0)}}function Ia(e,t,n){if(t!=null&&(t=""+On(t),t!==e.value&&(e.value=t),n==null)){e.defaultValue!==t&&(e.defaultValue=t);return}e.defaultValue=n!=null?""+On(n):""}function Xa(e,t,n,l){if(t==null){if(l!=null){if(n!=null)throw Error(u(92));if(mn(l)){if(1<l.length)throw Error(u(93));l=l[0]}n=l}n==null&&(n=""),t=n}n=On(t),e.defaultValue=n,l=e.textContent,l===n&&l!==""&&l!==null&&(e.value=l),qa(e)}function vo(e,t){if(t){var n=e.firstChild;if(n&&n===e.lastChild&&n.nodeType===3){n.nodeValue=t;return}}e.textContent=t}var qo=new Set("animationIterationCount aspectRatio borderImageOutset borderImageSlice borderImageWidth boxFlex boxFlexGroup boxOrdinalGroup columnCount columns flex flexGrow flexPositive flexShrink flexNegative flexOrder gridArea gridRow gridRowEnd gridRowSpan gridRowStart gridColumn gridColumnEnd gridColumnSpan gridColumnStart fontWeight lineClamp lineHeight opacity order orphans scale tabSize widows zIndex zoom fillOpacity floodOpacity stopOpacity strokeDasharray strokeDashoffset strokeMiterlimit strokeOpacity strokeWidth MozAnimationIterationCount MozBoxFlex MozBoxFlexGroup MozLineClamp msAnimationIterationCount msFlex msZoom msFlexGrow msFlexNegative msFlexOrder msFlexPositive msFlexShrink msGridColumn msGridColumnSpan msGridRow msGridRowSpan WebkitAnimationIterationCount WebkitBoxFlex WebKitBoxFlexGroup WebkitBoxOrdinalGroup WebkitColumnCount WebkitColumns WebkitFlex WebkitFlexGrow WebkitFlexPositive WebkitFlexShrink WebkitLineClamp".split(" "));function fl(e,t,n){var l=t.indexOf("--")===0;n==null||typeof n=="boolean"||n===""?l?e.setProperty(t,""):t==="float"?e.cssFloat="":e[t]="":l?e.setProperty(t,n):typeof n!="number"||n===0||qo.has(t)?t==="float"?e.cssFloat=n:e[t]=(""+n).trim():e[t]=n+"px"}function Qa(e,t,n){if(t!=null&&typeof t!="object")throw Error(u(62));if(e=e.style,n!=null){for(var l in n)!n.hasOwnProperty(l)||t!=null&&t.hasOwnProperty(l)||(l.indexOf("--")===0?e.setProperty(l,""):l==="float"?e.cssFloat="":e[l]="",yt=!0);for(var a in t)l=t[a],t.hasOwnProperty(a)&&n[a]!==l&&(fl(e,a,l),yt=!0)}else for(var i in t)t.hasOwnProperty(i)&&fl(e,i,t[i])}function _l(e){if(e.indexOf("-")===-1)return!1;switch(e){case"annotation-xml":case"color-profile":case"font-face":case"font-face-src":case"font-face-uri":case"font-face-format":case"font-face-name":case"missing-glyph":return!1;default:return!0}}var rd=new Map([["acceptCharset","accept-charset"],["htmlFor","for"],["httpEquiv","http-equiv"],["crossOrigin","crossorigin"],["accentHeight","accent-height"],["alignmentBaseline","alignment-baseline"],["arabicForm","arabic-form"],["baselineShift","baseline-shift"],["capHeight","cap-height"],["clipPath","clip-path"],["clipRule","clip-rule"],["colorInterpolation","color-interpolation"],["colorInterpolationFilters","color-interpolation-filters"],["colorProfile","color-profile"],["colorRendering","color-rendering"],["dominantBaseline","dominant-baseline"],["enableBackground","enable-background"],["fillOpacity","fill-opacity"],["fillRule","fill-rule"],["floodColor","flood-color"],["floodOpacity","flood-opacity"],["fontFamily","font-family"],["fontSize","font-size"],["fontSizeAdjust","font-size-adjust"],["fontStretch","font-stretch"],["fontStyle","font-style"],["fontVariant","font-variant"],["fontWeight","font-weight"],["glyphName","glyph-name"],["glyphOrientationHorizontal","glyph-orientation-horizontal"],["glyphOrientationVertical","glyph-orientation-vertical"],["horizAdvX","horiz-adv-x"],["horizOriginX","horiz-origin-x"],["imageRendering","image-rendering"],["letterSpacing","letter-spacing"],["lightingColor","lighting-color"],["markerEnd","marker-end"],["markerMid","marker-mid"],["markerStart","marker-start"],["maskType","mask-type"],["overlinePosition","overline-position"],["overlineThickness","overline-thickness"],["paintOrder","paint-order"],["panose-1","panose-1"],["pointerEvents","pointer-events"],["renderingIntent","rendering-intent"],["shapeRendering","shape-rendering"],["stopColor","stop-color"],["stopOpacity","stop-opacity"],["strikethroughPosition","strikethrough-position"],["strikethroughThickness","strikethrough-thickness"],["strokeDasharray","stroke-dasharray"],["strokeDashoffset","stroke-dashoffset"],["strokeLinecap","stroke-linecap"],["strokeLinejoin","stroke-linejoin"],["strokeMiterlimit","stroke-miterlimit"],["strokeOpacity","stroke-opacity"],["strokeWidth","stroke-width"],["textAnchor","text-anchor"],["textDecoration","text-decoration"],["textRendering","text-rendering"],["transformOrigin","transform-origin"],["underlinePosition","underline-position"],["underlineThickness","underline-thickness"],["unicodeBidi","unicode-bidi"],["unicodeRange","unicode-range"],["unitsPerEm","units-per-em"],["vAlphabetic","v-alphabetic"],["vHanging","v-hanging"],["vIdeographic","v-ideographic"],["vMathematical","v-mathematical"],["vectorEffect","vector-effect"],["vertAdvY","vert-adv-y"],["vertOriginX","vert-origin-x"],["vertOriginY","vert-origin-y"],["wordSpacing","word-spacing"],["writingMode","writing-mode"],["xmlnsXlink","xmlns:xlink"],["xHeight","x-height"]]),ei=/^[\u0000-\u001F ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:/i;function la(e){return ei.test(""+e)?"javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')":e}function no(){}var oa=null;function sa(e){return e=e.target||e.srcElement||window,e.correspondingUseElement&&(e=e.correspondingUseElement),e.nodeType===3?e.parentNode:e}var Io=null,zl=null;function Wn(e){var t=xo(e);if(t&&(e=t.stateNode)){var n=e[wn]||null;e:switch(e=t.stateNode,t.type){case"input":if(ms(e,n.value,n.defaultValue,n.defaultValue,n.checked,n.defaultChecked,n.type,n.name),t=n.name,n.type==="radio"&&t!=null){for(n=e;n.parentNode;)n=n.parentNode;for(n=n.querySelectorAll('input[name="'+Xn(""+t)+'"][type="radio"]'),t=0;t<n.length;t++){var l=n[t];if(l!==e&&l.form===e.form){var a=l[wn]||null;if(!a)throw Error(u(90));ms(l,a.value,a.defaultValue,a.defaultValue,a.checked,a.defaultChecked,a.type,a.name)}}for(t=0;t<n.length;t++)l=n[t],l.form===e.form&&Pr(l)}break e;case"textarea":Ia(e,n.value,n.defaultValue);break e;case"select":t=n.value,t!=null&&Yo(e,!!n.multiple,t,!1)}}}var Vl=!1;function hs(e,t,n){if(Vl)return e(t,n);Vl=!0;try{var l=e(t);return l}finally{if(Vl=!1,(Io!==null||zl!==null)&&(pu(),Io&&(t=Io,e=zl,zl=Io=null,Wn(t),e)))for(t=0;t<e.length;t++)Wn(e[t])}}function Sn(e,t){var n=e.stateNode;if(n===null)return null;var l=n[wn]||null;if(l===null)return null;n=l[t];e:switch(t){case"onClick":case"onClickCapture":case"onDoubleClick":case"onDoubleClickCapture":case"onMouseDown":case"onMouseDownCapture":case"onMouseMove":case"onMouseMoveCapture":case"onMouseUp":case"onMouseUpCapture":case"onMouseEnter":(l=!l.disabled)||(e=e.type,l=!(e==="button"||e==="input"||e==="select"||e==="textarea")),e=!l;break e;default:e=!1}if(e)return null;if(n&&typeof n!="function")throw Error(u(231,t,typeof n));return n}var lo=!(typeof window>"u"||typeof window.document>"u"||typeof window.document.createElement>"u"),Ga=!1;if(lo)try{var gs={};Object.defineProperty(gs,"passive",{get:function(){Ga=!0}}),window.addEventListener("test",gs,gs),window.removeEventListener("test",gs,gs)}catch{Ga=!1}var wt=null,ti=null,ml=null;function ni(){if(ml)return ml;var e,t=ti,n=t.length,l,a="value"in wt?wt.value:wt.textContent,i=a.length;for(e=0;e<n&&t[e]===a[e];e++);var _=n-e;for(l=1;l<=_&&t[n-l]===a[i-l];l++);return ml=a.slice(e,1<l?1-l:void 0)}function aa(e){var t=e.keyCode;return"charCode"in e?(e=e.charCode,e===0&&t===13&&(e=13)):e=t,e===10&&(e=13),32<=e||e===13?e:0}function ra(){return!0}function pc(){return!1}function Kn(e){function t(n,l,a,i,_){this._reactName=n,this._targetInst=a,this.type=l,this.nativeEvent=i,this.target=_,this.currentTarget=null;for(var y in e)e.hasOwnProperty(y)&&(n=e[y],this[y]=n?n(i):i[y]);return this.isDefaultPrevented=(i.defaultPrevented!=null?i.defaultPrevented:i.returnValue===!1)?ra:pc,this.isPropagationStopped=pc,this}return W(t.prototype,{preventDefault:function(){this.defaultPrevented=!0;var n=this.nativeEvent;n&&(n.preventDefault?n.preventDefault():typeof n.returnValue!="unknown"&&(n.returnValue=!1),this.isDefaultPrevented=ra)},stopPropagation:function(){var n=this.nativeEvent;n&&(n.stopPropagation?n.stopPropagation():typeof n.cancelBubble!="unknown"&&(n.cancelBubble=!0),this.isPropagationStopped=ra)},persist:function(){},isPersistent:ra}),t}var wo={eventPhase:0,bubbles:0,cancelable:0,timeStamp:function(e){return e.timeStamp||Date.now()},defaultPrevented:0,isTrusted:0},Sl=Kn(wo),dn=W({},wo,{view:0,detail:0}),li=Kn(dn),Va,oo,Al,ln=W({},dn,{screenX:0,screenY:0,clientX:0,clientY:0,pageX:0,pageY:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,getModifierState:ua,button:0,buttons:0,relatedTarget:function(e){return e.relatedTarget===void 0?e.fromElement===e.srcElement?e.toElement:e.fromElement:e.relatedTarget},movementX:function(e){return"movementX"in e?e.movementX:(e!==Al&&(Al&&e.type==="mousemove"?(Va=e.screenX-Al.screenX,oo=e.screenY-Al.screenY):oo=Va=0,Al=e),Va)},movementY:function(e){return"movementY"in e?e.movementY:oo}}),Fa=Kn(ln),Wa=W({},ln,{dataTransfer:0}),yc=Kn(Wa),ia=W({},dn,{relatedTarget:0}),ps=Kn(ia),Xo=W({},wo,{animationName:0,elapsedTime:0,pseudoElement:0}),Ka=Kn(Xo),xc=W({},wo,{clipboardData:function(e){return"clipboardData"in e?e.clipboardData:window.clipboardData}}),So=Kn(xc),Za=W({},wo,{data:0}),hl=Kn(Za),ca={Esc:"Escape",Spacebar:" ",Left:"ArrowLeft",Up:"ArrowUp",Right:"ArrowRight",Down:"ArrowDown",Del:"Delete",Win:"OS",Menu:"ContextMenu",Apps:"ContextMenu",Scroll:"ScrollLock",MozPrintableKey:"Unidentified"},Qo={8:"Backspace",9:"Tab",12:"Clear",13:"Enter",16:"Shift",17:"Control",18:"Alt",19:"Pause",20:"CapsLock",27:"Escape",32:" ",33:"PageUp",34:"PageDown",35:"End",36:"Home",37:"ArrowLeft",38:"ArrowUp",39:"ArrowRight",40:"ArrowDown",45:"Insert",46:"Delete",112:"F1",113:"F2",114:"F3",115:"F4",116:"F5",117:"F6",118:"F7",119:"F8",120:"F9",121:"F10",122:"F11",123:"F12",144:"NumLock",145:"ScrollLock",224:"Meta"},ys={Alt:"altKey",Control:"ctrlKey",Meta:"metaKey",Shift:"shiftKey"};function so(e){var t=this.nativeEvent;return t.getModifierState?t.getModifierState(e):(e=ys[e])?!!t[e]:!1}function ua(){return so}var Go=W({},dn,{key:function(e){if(e.key){var t=ca[e.key]||e.key;if(t!=="Unidentified")return t}return e.type==="keypress"?(e=aa(e),e===13?"Enter":String.fromCharCode(e)):e.type==="keydown"||e.type==="keyup"?Qo[e.keyCode]||"Unidentified":""},code:0,location:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,repeat:0,locale:0,getModifierState:ua,charCode:function(e){return e.type==="keypress"?aa(e):0},keyCode:function(e){return e.type==="keydown"||e.type==="keyup"?e.keyCode:0},which:function(e){return e.type==="keypress"?aa(e):e.type==="keydown"||e.type==="keyup"?e.keyCode:0}}),jl=Kn(Go),da=W({},ln,{pointerId:0,width:0,height:0,pressure:0,tangentialPressure:0,tiltX:0,tiltY:0,twist:0,pointerType:0,isPrimary:0}),Vo=Kn(da),fa=W({},wo,{submitter:0}),bc=Kn(fa),Ol=W({},dn,{touches:0,targetTouches:0,changedTouches:0,altKey:0,metaKey:0,ctrlKey:0,shiftKey:0,getModifierState:ua}),vc=Kn(Ol),oi=W({},wo,{propertyName:0,elapsedTime:0,pseudoElement:0}),id=Kn(oi),si=W({},ln,{deltaX:function(e){return"deltaX"in e?e.deltaX:"wheelDeltaX"in e?-e.wheelDeltaX:0},deltaY:function(e){return"deltaY"in e?e.deltaY:"wheelDeltaY"in e?-e.wheelDeltaY:"wheelDelta"in e?-e.wheelDelta:0},deltaZ:0,deltaMode:0}),Ja=Kn(si),_a=W({},wo,{newState:0,oldState:0,source:0}),cd=Kn(_a),Pa=[9,13,27,32],er=lo&&"CompositionEvent"in window,jo=null;lo&&"documentMode"in document&&(jo=document.documentMode);var wc=lo&&"TextEvent"in window&&!jo,tr=lo&&(!er||jo&&8<jo&&11>=jo),Sc=" ",nr=!1;function lr(e,t){switch(e){case"keyup":return Pa.indexOf(t.keyCode)!==-1;case"keydown":return t.keyCode!==229;case"keypress":case"mousedown":case"focusout":return!0;default:return!1}}function xs(e){return e=e.detail,typeof e=="object"&&"data"in e?e.data:null}var ao=!1;function jc(e,t){switch(e){case"compositionend":return xs(t);case"keypress":return t.which!==32?null:(nr=!0,Sc);case"textInput":return e=t.data,e===Sc&&nr?null:e;default:return null}}function or(e,t){if(ao)return e==="compositionend"||!er&&lr(e,t)?(e=ni(),ml=ti=wt=null,ao=!1,e):null;switch(e){case"paste":return null;case"keypress":if(!(t.ctrlKey||t.altKey||t.metaKey)||t.ctrlKey&&t.altKey){if(t.char&&1<t.char.length)return t.char;if(t.which)return String.fromCharCode(t.which)}return null;case"compositionend":return tr&&t.locale!=="ko"?null:t.data;default:return null}}var ai={color:!0,date:!0,datetime:!0,"datetime-local":!0,email:!0,month:!0,number:!0,password:!0,range:!0,search:!0,tel:!0,text:!0,time:!0,url:!0,week:!0};function bs(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t==="input"?!!ai[e.type]:t==="textarea"}function ma(e,t,n,l){Io?zl?zl.push(l):zl=[l]:Io=l,t=Su(t,"onChange"),0<t.length&&(n=new Sl("onChange","change",null,n,l),e.push({event:n,listeners:t}))}var al=null,ha=null;function ri(e){vg(e,0)}function vs(e){var t=cs(e);if(Pr(t))return e}function sr(e,t){if(e==="change")return t}var ga=!1;if(lo){var ii;if(lo){var ci="oninput"in document;if(!ci){var kc=document.createElement("div");kc.setAttribute("oninput","return;"),ci=typeof kc.oninput=="function"}ii=ci}else ii=!1;ga=ii&&(!document.documentMode||9<document.documentMode)}function pa(){al&&(al.detachEvent("onpropertychange",ko),ha=al=null)}function ko(e){if(e.propertyName==="value"&&vs(ha)){var t=[];ma(t,ha,e,sa(e)),hs(ri,t)}}function ui(e,t,n){e==="focusin"?(pa(),al=t,ha=n,al.attachEvent("onpropertychange",ko)):e==="focusout"&&pa()}function di(e){if(e==="selectionchange"||e==="keyup"||e==="keydown")return vs(ha)}function Fl(e,t){if(e==="click")return vs(t)}function ar(e,t){if(e==="input"||e==="change")return vs(t)}function fi(e,t){return e===t&&(e!==0||1/e===1/t)||e!==e&&t!==t}var gl=typeof Object.is=="function"?Object.is:fi;function ro(e,t){if(gl(e,t))return!0;if(typeof e!="object"||e===null||typeof t!="object"||t===null)return!1;var n=Object.keys(e),l=Object.keys(t);if(n.length!==l.length)return!1;for(l=0;l<n.length;l++){var a=n[l];if(!_e.call(t,a)||!gl(e[a],t[a]))return!1}return!0}function ws(e){if(e=e||(typeof document<"u"?document:void 0),typeof e>"u")return null;try{return e.activeElement||e.body}catch{return e.body}}function ya(e){for(;e&&e.firstChild;)e=e.firstChild;return e}function _i(e,t){var n=ya(e);e=0;for(var l;n;){if(n.nodeType===3){if(l=e+n.textContent.length,e<=t&&l>=t)return{node:n,offset:t-e};e=l}e:{for(;n;){if(n.nextSibling){n=n.nextSibling;break e}n=n.parentNode}n=void 0}n=ya(n)}}function mi(e,t){return e&&t?e===t?!0:e&&e.nodeType===3?!1:t&&t.nodeType===3?mi(e,t.parentNode):"contains"in e?e.contains(t):e.compareDocumentPosition?!!(e.compareDocumentPosition(t)&16):!1:!1}function hi(e){e=e!=null&&e.ownerDocument!=null&&e.ownerDocument.defaultView!=null?e.ownerDocument.defaultView:window;for(var t=ws(e.document);t instanceof e.HTMLIFrameElement;){try{var n=typeof t.contentWindow.location.href=="string"}catch{n=!1}if(n)e=t.contentWindow;else break;t=ws(e.document)}return t}function rr(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t&&(t==="input"&&(e.type==="text"||e.type==="search"||e.type==="tel"||e.type==="url"||e.type==="password")||t==="textarea"||e.contentEditable==="true")}var Cc=lo&&"documentMode"in document&&11>=document.documentMode,Fo=null,rl=null,Ss=null,ir=!1;function gi(e,t,n){var l=n.window===n?n.document:n.nodeType===9?n:n.ownerDocument;ir||Fo==null||Fo!==ws(l)||(l=Fo,"selectionStart"in l&&rr(l)?l={start:l.selectionStart,end:l.selectionEnd}:(l=(l.ownerDocument&&l.ownerDocument.defaultView||window).getSelection(),l={anchorNode:l.anchorNode,anchorOffset:l.anchorOffset,focusNode:l.focusNode,focusOffset:l.focusOffset}),Ss&&ro(Ss,l)||(Ss=l,l=Su(rl,"onSelect"),0<l.length&&(t=new Sl("onSelect","select",null,t,n),e.push({event:t,listeners:l}),t.target=Fo)))}function js(e,t){var n={};return n[e.toLowerCase()]=t.toLowerCase(),n["Webkit"+e]="webkit"+t,n["Moz"+e]="moz"+t,n}var Wo={animationend:js("Animation","AnimationEnd"),animationiteration:js("Animation","AnimationIteration"),animationstart:js("Animation","AnimationStart"),transitionrun:js("Transition","TransitionRun"),transitionstart:js("Transition","TransitionStart"),transitioncancel:js("Transition","TransitionCancel"),transitionend:js("Transition","TransitionEnd")},cr={},Ko={};lo&&(Ko=document.createElement("div").style,"AnimationEvent"in window||(delete Wo.animationend.animation,delete Wo.animationiteration.animation,delete Wo.animationstart.animation),"TransitionEvent"in window||delete Wo.transitionend.transition);function Co(e){if(cr[e])return cr[e];if(!Wo[e])return e;var t=Wo[e],n;for(n in t)if(t.hasOwnProperty(n)&&n in Ko)return cr[e]=t[n];return e}var pi=Co("animationend"),yi=Co("animationiteration"),xi=Co("animationstart"),Nc=Co("transitionrun"),Mc=Co("transitionstart"),Ec=Co("transitioncancel"),x=Co("transitionend"),S=new Map,N="abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error fullscreenChange fullscreenError gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");N.push("scrollEnd");function T(e,t){S.set(e,t),Ho(t,[e])}var Q=0;function H(e,t){if(e.name!=null&&e.name!=="auto")return e.name;if(t.autoName!==null)return t.autoName;e=fo.identifierPrefix;var n=Q++;return e="_"+e+"t_"+n.toString(32)+"_",t.autoName=e}function Y(e){if(e==null||typeof e=="string")return e;var t=null,n=kr;if(n!==null)for(var l=0;l<n.length;l++){var a=e[n[l]];if(a!=null){if(a==="none")return"none";t=t==null?a:t+(" "+a)}}return t??e.default}function ce(e,t){return e=Y(e),t=Y(t),t==null?e==="auto"?null:e:t==="auto"?null:t}var me=typeof reportError=="function"?reportError:function(e){if(typeof window=="object"&&typeof window.ErrorEvent=="function"){var t=new window.ErrorEvent("error",{bubbles:!0,cancelable:!0,message:typeof e=="object"&&e!==null&&typeof e.message=="string"?String(e.message):String(e),error:e});if(!window.dispatchEvent(t))return}else if(typeof process=="object"&&typeof process.emit=="function"){process.emit("uncaughtException",e);return}console.error(e)},ge=[],Te=0,qe=0;function Pe(){for(var e=Te,t=qe=Te=0;t<e;){var n=ge[t];ge[t++]=null;var l=ge[t];ge[t++]=null;var a=ge[t];ge[t++]=null;var i=ge[t];if(ge[t++]=null,l!==null&&a!==null){var _=l.pending;_===null?a.next=a:(a.next=_.next,_.next=a),l.pending=a}i!==0&&st(n,a,i)}}function dt(e,t,n,l){ge[Te++]=e,ge[Te++]=t,ge[Te++]=n,ge[Te++]=l,qe|=l,e.lanes|=l,e=e.alternate,e!==null&&(e.lanes|=l)}function $e(e,t,n,l){return dt(e,t,n,l),St(e)}function Ve(e,t){return dt(e,null,null,t),St(e)}function st(e,t,n){e.lanes|=n;var l=e.alternate;l!==null&&(l.lanes|=n);for(var a=!1,i=e.return;i!==null;)i.childLanes|=n,l=i.alternate,l!==null&&(l.childLanes|=n),i.tag===22&&(e=i.stateNode,e===null||e._visibility&1||(a=!0)),e=i,i=i.return;return e.tag===3?(i=e.stateNode,a&&t!==null&&(a=31-Gt(n),e=i.hiddenUpdates,l=e[a],l===null?e[a]=[t]:l.push(t),t.lane=n|536870912),i):null}function St(e){if(50<Hi)throw Hi=0,gu=null,Error(u(185));for(var t=e.return;t!==null;)e=t,t=e.return;return e.tag===3?e.stateNode:null}var $t={};function Rn(e,t,n,l){this.tag=e,this.key=n,this.sibling=this.child=this.return=this.stateNode=this.type=this.elementType=null,this.index=0,this.refCleanup=this.ref=null,this.pendingProps=t,this.dependencies=this.memoizedState=this.updateQueue=this.memoizedProps=null,this.mode=l,this.subtreeFlags=this.flags=0,this.deletions=null,this.childLanes=this.lanes=0,this.alternate=null}function He(e,t,n,l){return new Rn(e,t,n,l)}function et(e){return e=e.prototype,!(!e||!e.isReactComponent)}function Ct(e,t){var n=e.alternate;return n===null?(n=He(e.tag,t,e.key,e.mode),n.elementType=e.elementType,n.type=e.type,n.stateNode=e.stateNode,n.alternate=e,e.alternate=n):(n.pendingProps=t,n.type=e.type,n.flags=0,n.subtreeFlags=0,n.deletions=null),n.flags=e.flags&1206910976,n.childLanes=e.childLanes,n.lanes=e.lanes,n.child=e.child,n.memoizedProps=e.memoizedProps,n.memoizedState=e.memoizedState,n.updateQueue=e.updateQueue,t=e.dependencies,n.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext},n.sibling=e.sibling,n.index=e.index,n.ref=e.ref,n.refCleanup=e.refCleanup,n}function Wt(e,t){e.flags&=1206910978;var n=e.alternate;return n===null?(e.childLanes=0,e.lanes=t,e.child=null,e.subtreeFlags=0,e.memoizedProps=null,e.memoizedState=null,e.updateQueue=null,e.dependencies=null,e.stateNode=null):(e.childLanes=n.childLanes,e.lanes=n.lanes,e.child=n.child,e.subtreeFlags=0,e.deletions=null,e.memoizedProps=n.memoizedProps,e.memoizedState=n.memoizedState,e.updateQueue=n.updateQueue,e.type=n.type,t=n.dependencies,e.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext}),e}function il(e,t,n,l,a,i){var _=0;if(l=e,typeof l=="function")et(l)&&(_=1);else if(typeof l=="string")_=Yx(e,n,Ge.current)?26:e==="html"||e==="head"||e==="body"?27:5;else e:switch(l){case Qe:return e=He(31,n,t,a),e.elementType=Qe,e.lanes=i,e;case pe:return fn(n.children,a,i,t);case V:_=8,a|=24;break;case Ie:return e=He(12,n,t,a|2),e.elementType=Ie,e.lanes=i,e;case we:return e=He(13,n,t,a),e.elementType=we,e.lanes=i,e;case fe:return e=He(19,n,t,a),e.elementType=fe,e.lanes=i,e;case Ut:case Ne:return e=a|32,e=He(30,n,t,e),e.elementType=Ne,e.lanes=i,e.stateNode={autoName:null,paired:null,clones:null,ref:null},e;default:if(typeof l=="object"&&l!==null)switch(l.$$typeof){case ue:_=10;break e;case he:_=9;break e;case G:_=11;break e;case Oe:_=14;break e;case Se:_=16,l=null;break e}_=29,n=Error(u(130,e===null?"null":typeof e,"")),l=null}return t=He(_,n,t,a),t.elementType=e,t.type=l,t.lanes=i,t}function fn(e,t,n,l){return e=He(7,e,l,t),e.lanes=n,e}function No(e,t,n){return e=He(6,e,null,t),e.lanes=n,e}function Zo(e){var t=He(18,null,null,0);return t.stateNode=e,t}function zn(e,t,n){return t=He(4,e.children!==null?e.children:[],e.key,t),t.lanes=n,t.stateNode={containerInfo:e.containerInfo,pendingChildren:null,implementation:e.implementation},t}var ks=new WeakMap;function Zn(e,t){if(typeof e=="object"&&e!==null){var n=ks.get(e);return n!==void 0?n:(t={value:e,source:t,stack:ke(t)},ks.set(e,t),t)}return{value:e,source:t,stack:ke(t)}}var Dl=[],on=0,Ll=null,Bl=0,kl=[],Cl=0,Mo=null,Wl=1,Eo="";function Jo(e,t){Dl[on++]=Bl,Dl[on++]=Ll,Ll=e,Bl=t}function nm(e,t,n){kl[Cl++]=Wl,kl[Cl++]=Eo,kl[Cl++]=Mo,Mo=e;var l=Wl;e=Eo;var a=32-Gt(l)-1;l&=~(1<<a),n+=1;var i=32-Gt(t)+a;if(30<i){var _=a-a%5;i=(l&(1<<_)-1).toString(32),l>>=_,a-=_,Wl=1<<32-Gt(t)+a|n<<a|l,Eo=i+e}else Wl=1<<i|n<<a|l,Eo=e}function Tc(e){e.return!==null&&(Jo(e,1),nm(e,1,0))}function ud(e){for(;e===Ll;)Ll=Dl[--on],Dl[on]=null,Bl=Dl[--on],Dl[on]=null;for(;e===Mo;)Mo=kl[--Cl],kl[Cl]=null,Eo=kl[--Cl],kl[Cl]=null,Wl=kl[--Cl],kl[Cl]=null}function lm(e,t){kl[Cl++]=Wl,kl[Cl++]=Eo,kl[Cl++]=Mo,Wl=t.id,Eo=t.overflow,Mo=e}var Qn=null,rn=null,jt=!1,Cs=null,Kl=!1,dd=Error(u(519));function Ns(e){var t=Error(u(418,1<arguments.length&&arguments[1]!==void 0&&arguments[1]?"text":"HTML",""));throw bi(Zn(t,e)),dd}function om(e){var t=e.stateNode,n=e.type,l=e.memoizedProps;switch(t[En]=e,t[wn]=l,n){case"dialog":Et("cancel",t),Et("close",t);break;case"iframe":case"object":case"embed":Et("load",t);break;case"video":case"audio":for(n=0;n<qi.length;n++)Et(qi[n],t);break;case"source":Et("error",t);break;case"img":case"image":case"link":Et("error",t),Et("load",t);break;case"details":Et("toggle",t);break;case"input":Et("invalid",t),wl(t,l.value,l.defaultValue,l.checked,l.defaultChecked,l.type,l.name,!0);break;case"select":Et("invalid",t);break;case"textarea":Et("invalid",t),Xa(t,l.value,l.defaultValue,l.children)}n=l.children,typeof n!="string"&&typeof n!="number"&&typeof n!="bigint"||t.textContent===""+n||l.suppressHydrationWarning===!0||kg(t.textContent,n)?(l.popover!=null&&(Et("beforetoggle",t),Et("toggle",t)),l.onScroll!=null&&Et("scroll",t),l.onScrollEnd!=null&&Et("scrollend",t),l.onClick!=null&&(t.onclick=no),t=!0):t=!1,t||Ns(e,!0)}function Rc(e){for(Qn=e.return;Qn;)switch(Qn.tag){case 5:case 31:case 13:Kl=!1;return;case 27:case 3:Kl=!0;return;default:Qn=Qn.return}}function ur(e){if(e!==Qn)return!1;if(!jt)return Rc(e),jt=!0,!1;var t=e.tag,n;if((n=t!==3&&t!==27)&&((n=t===5)&&(n=e.type,n=!(n!=="form"&&n!=="button")||qf(e.type,e.memoizedProps)),n=!n),n&&rn&&Ns(e),Rc(e),t===13){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(u(317));rn=Xg(e)}else if(t===31){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(u(317));rn=Xg(e)}else t===27?(t=rn,Is(e.type)?(e=Zf,Zf=null,rn=e):rn=t):rn=Qn?Jl(e.stateNode.nextSibling):null;return!0}function xa(){rn=Qn=null,jt=!1}function fd(){var e=Cs;return e!==null&&(El===null?El=e:El.push.apply(El,e),Cs=null),e}function bi(e){Cs===null?Cs=[e]:Cs.push(e)}var _d=bn(null),ba=null,Po=null;function Ms(e,t,n){ye(_d,t._currentValue),t._currentValue=n}function es(e){e._currentValue=_d.current,re(_d)}function zc(e,t,n){for(;e!==null;){var l=e.alternate;if((e.childLanes&t)!==t?(e.childLanes|=t,l!==null&&(l.childLanes|=t)):l!==null&&(l.childLanes&t)!==t&&(l.childLanes|=t),e===n)break;e=e.return}}function md(e,t,n,l){var a=e.child;for(a!==null&&(a.return=e);a!==null;){var i=a.dependencies;if(i!==null){var _=a.child;i=i.firstContext;e:for(;i!==null;){var y=i;i=a;for(var M=0;M<t.length;M++)if(y.context===t[M]){i.lanes|=n,y=i.alternate,y!==null&&(y.lanes|=n),zc(i.return,n,e),l||(_=null);break e}i=y.next}}else if(a.tag===18){if(_=a.return,_===null)throw Error(u(341));_.lanes|=n,i=_.alternate,i!==null&&(i.lanes|=n),zc(_,n,e),_=null}else a.tag===13&&a.memoizedState!==null&&a.memoizedState.dehydrated===null?(a.lanes|=n,_=a.alternate,_!==null&&(_.lanes|=n),zc(a.return,n,e),_=a.child,_=_!==null?_.sibling:null):_=a.child;if(_!==null)_.return=a;else for(_=a;_!==null;){if(_===e){_=null;break}if(a=_.sibling,a!==null){a.return=_.return,_=a;break}_=_.return}a=_}}function va(e,t,n,l){e=null;for(var a=t,i=!1;a!==null;){if(!i){if((a.flags&524288)!==0)i=!0;else if((a.flags&262144)!==0)break}if(a.tag===10){var _=a.alternate;if(_===null)throw Error(u(387));if(_=_.memoizedProps,_!==null){var y=a.type;gl(a.pendingProps.value,_.value)||(e!==null?e.push(y):e=[y])}}else if(a===rt.current){if(_=a.alternate,_===null)throw Error(u(387));_.memoizedState.memoizedState!==a.memoizedState.memoizedState&&(e!==null?e.push(Dr):e=[Dr])}a=a.return}return e!==null&&md(t,e,n,l),t.flags|=262144,e!==null}function Ac(e){for(e=e.firstContext;e!==null;){if(!gl(e.context._currentValue,e.memoizedValue))return!0;e=e.next}return!1}function wa(e){ba=e,Po=null,e=e.dependencies,e!==null&&(e.firstContext=null)}function Jn(e){return sm(ba,e)}function Oc(e,t){return ba===null&&wa(e),sm(e,t)}function sm(e,t){var n=t._currentValue;if(t={context:t,memoizedValue:n,next:null},Po===null){if(e===null)throw Error(u(308));Po=t,e.dependencies={lanes:0,firstContext:t},e.flags|=524288}else Po=Po.next=t;return n}var my=typeof AbortController<"u"?AbortController:function(){var e=[],t=this.signal={aborted:!1,addEventListener:function(n,l){e.push(l)}};this.abort=function(){t.aborted=!0,e.forEach(function(n){return n()})}},hy=s.unstable_scheduleCallback,gy=s.unstable_NormalPriority,Dn={$$typeof:ue,Consumer:null,Provider:null,_currentValue:null,_currentValue2:null,_threadCount:0};function hd(){return{controller:new my,data:new Map,refCount:0}}function vi(e){e.refCount--,e.refCount===0&&hy(gy,function(){e.controller.abort()})}function am(e,t){if((e.pendingLanes&4194048)!==0){var n=e.transitionTypes;for(n===null&&(n=e.transitionTypes=[]),e=0;e<t.length;e++){var l=t[e];n.indexOf(l)===-1&&n.push(l)}}}var wi=null;function py(e){var t=e.transitionTypes;return e.transitionTypes=null,t}var Si=null,gd=0,Sa=0,dr=null;function yy(e,t){if(Si===null){var n=Si=[];gd=0,Sa=Af(),dr={status:"pending",value:void 0,then:function(l){n.push(l)}}}return gd++,t.then(rm,rm),t}function rm(){if(--gd===0&&(wi=null,Si!==null)){dr!==null&&(dr.status="fulfilled");var e=Si;Si=null,Sa=0,dr=null;for(var t=0;t<e.length;t++)(0,e[t])()}}function xy(e,t){var n=[],l={status:"pending",value:null,reason:null,then:function(a){n.push(a)}};return e.then(function(){l.status="fulfilled",l.value=t;for(var a=0;a<n.length;a++)(0,n[a])(t)},function(a){for(l.status="rejected",l.reason=a,a=0;a<n.length;a++)(0,n[a])(void 0)}),l}var im=Re.S;Re.S=function(e,t){if(eg=Ze(),typeof t=="object"&&t!==null&&typeof t.then=="function"&&yy(e,t),wi!==null)for(var n=Er;n!==null;)am(n,wi),n=n.next;if(n=e.types,n!==null){for(var l=Er;l!==null;)am(l,n),l=l.next;if(Sa!==0){l=wi,l===null&&(l=wi=[]);for(var a=0;a<n.length;a++){var i=n[a];l.indexOf(i)===-1&&l.push(i)}}}im!==null&&im(e,t)};var ja=bn(null);function pd(){var e=ja.current;return e!==null?e:sn.pooledCache}function Dc(e,t){t===null?ye(ja,ja.current):ye(ja,t.pool)}function cm(){var e=pd();return e===null?null:{parent:Dn._currentValue,pool:e}}var fr=Error(u(460)),yd=Error(u(474)),Lc=Error(u(542)),Bc={then:function(){}};function um(e){return e=e.status,e==="fulfilled"||e==="rejected"}function dm(e,t,n){switch(n=e[n],n===void 0?e.push(t):n!==t&&(t.then(no,no),t=n),t.status){case"fulfilled":return t.value;case"rejected":throw e=t.reason,_m(e),e===void 0&&!("reason"in t)?Error(u(600)):e;default:if(typeof t.status=="string")t.then(no,no);else{if(e=sn,e!==null&&100<e.shellSuspendCounter)throw Error(u(482));e=t,e.status="pending",e.then(function(l){if(t.status==="pending"){var a=t;a.status="fulfilled",a.value=l}},function(l){if(t.status==="pending"){var a=t;a.status="rejected",a.reason=l}})}switch(t.status){case"fulfilled":return t.value;case"rejected":throw e=t.reason,_m(e),e}throw Ca=t,fr}}function ka(e){try{var t=e._init;return t(e._payload)}catch(n){throw n!==null&&typeof n=="object"&&typeof n.then=="function"?(Ca=n,fr):n}}var Ca=null;function fm(){if(Ca===null)throw Error(u(459));var e=Ca;return Ca=null,e}function _m(e){if(e===fr||e===Lc)throw Error(u(483))}var _r=null,ji=0;function $c(e){var t=ji;return ji+=1,_r===null&&(_r=[]),dm(_r,e,t)}function Es(e,t){t=t.props.ref,e.ref=t!==void 0?t:null}function Uc(e,t){throw t.$$typeof===Ee?Error(u(525)):(e=Object.prototype.toString.call(t),Error(u(31,e==="[object Object]"?"object with keys {"+Object.keys(t).join(", ")+"}":e)))}function mm(e){function t(D,R){if(e){var U=D.deletions;U===null?(D.deletions=[R],D.flags|=16):U.push(R)}}function n(D,R){if(!e)return null;for(;R!==null;)t(D,R),R=R.sibling;return null}function l(D){for(var R=new Map;D!==null;)D.key===null?R.set(D.index,D):R.set(D.key,D),D=D.sibling;return R}function a(D,R){return D=Ct(D,R),D.index=0,D.sibling=null,D}function i(D,R,U){return D.index=U,e?(U=D.alternate,U!==null?(U=U.index,U<R?(D.flags|=2,R):U):(D.flags|=134217730,R)):(D.flags|=1048576,R)}function _(D){return e&&D.alternate===null&&(D.flags|=134217730),D}function y(D,R,U,Z){return R===null||R.tag!==6?(R=No(U,D.mode,Z),R.return=D,R):(R=a(R,U),R.return=D,R)}function M(D,R,U,Z){var ze=U.type;return ze===pe?(D=X(D,R,U.props.children,Z,U.key),Es(D,U),D):R!==null&&(R.elementType===ze||typeof ze=="object"&&ze!==null&&ze.$$typeof===Se&&ka(ze)===R.type)?(R=a(R,U.props),Es(R,U),R.return=D,R):(R=il(U.type,U.key,U.props,null,D.mode,Z),Es(R,U),R.return=D,R)}function L(D,R,U,Z){return R===null||R.tag!==4||R.stateNode.containerInfo!==U.containerInfo||R.stateNode.implementation!==U.implementation?(R=zn(U,D.mode,Z),R.return=D,R):(R=a(R,U.children||[]),R.return=D,R)}function X(D,R,U,Z,ze){return R===null||R.tag!==7?(R=fn(U,D.mode,Z,ze),R.return=D,R):(R=a(R,U),R.return=D,R)}function J(D,R,U){if(typeof R=="string"&&R!==""||typeof R=="number"||typeof R=="bigint")return R=No(""+R,D.mode,U),R.return=D,R;if(typeof R=="object"&&R!==null){switch(R.$$typeof){case ne:return U=il(R.type,R.key,R.props,null,D.mode,U),Es(U,R),U.return=D,U;case Ce:return R=zn(R,D.mode,U),R.return=D,R;case Se:return R=ka(R),J(D,R,U)}if(mn(R)||We(R))return R=fn(R,D.mode,U,null),R.return=D,R;if(typeof R.then=="function")return J(D,$c(R),U);if(R.$$typeof===ue)return J(D,Oc(D,R),U);Uc(D,R)}return null}function A(D,R,U,Z){var ze=R!==null?R.key:null;if(typeof U=="string"&&U!==""||typeof U=="number"||typeof U=="bigint")return ze!==null?null:y(D,R,""+U,Z);if(typeof U=="object"&&U!==null){switch(U.$$typeof){case ne:return U.key===ze?M(D,R,U,Z):null;case Ce:return U.key===ze?L(D,R,U,Z):null;case Se:return U=ka(U),A(D,R,U,Z)}if(mn(U)||We(U))return ze!==null?null:X(D,R,U,Z,null);if(typeof U.then=="function")return A(D,R,$c(U),Z);if(U.$$typeof===ue)return A(D,R,Oc(D,U),Z);Uc(D,U)}return null}function q(D,R,U,Z,ze){if(typeof Z=="string"&&Z!==""||typeof Z=="number"||typeof Z=="bigint")return D=D.get(U)||null,y(R,D,""+Z,ze);if(typeof Z=="object"&&Z!==null){switch(Z.$$typeof){case ne:return D=D.get(Z.key===null?U:Z.key)||null,M(R,D,Z,ze);case Ce:return D=D.get(Z.key===null?U:Z.key)||null,L(R,D,Z,ze);case Se:return Z=ka(Z),q(D,R,U,Z,ze)}if(mn(Z)||We(Z))return D=D.get(U)||null,X(R,D,Z,ze,null);if(typeof Z.then=="function")return q(D,R,U,$c(Z),ze);if(Z.$$typeof===ue)return q(D,R,U,Oc(R,Z),ze);Uc(R,Z)}return null}function ve(D,R,U,Z){for(var ze=null,zt=null,Xe=R,tt=R=0,$n=null;Xe!==null&&tt<U.length;tt++){Xe.index>tt?($n=Xe,Xe=null):$n=Xe.sibling;var Lt=A(D,Xe,U[tt],Z);if(Lt===null){Xe===null&&(Xe=$n);break}e&&Xe&&Lt.alternate===null&&t(D,Xe),R=i(Lt,R,tt),zt===null?ze=Lt:zt.sibling=Lt,zt=Lt,Xe=$n}if(tt===U.length)return n(D,Xe),jt&&Jo(D,tt),ze;if(Xe===null){for(;tt<U.length;tt++)Xe=J(D,U[tt],Z),Xe!==null&&(R=i(Xe,R,tt),zt===null?ze=Xe:zt.sibling=Xe,zt=Xe);return jt&&Jo(D,tt),ze}for(Xe=l(Xe);tt<U.length;tt++)$n=q(Xe,D,tt,U[tt],Z),$n!==null&&(e&&(Lt=$n.alternate,Lt!==null&&Xe.delete(Lt.key===null?tt:Lt.key)),R=i($n,R,tt),zt===null?ze=$n:zt.sibling=$n,zt=$n);return e&&Xe.forEach(function(Fs){return t(D,Fs)}),jt&&Jo(D,tt),ze}function Le(D,R,U,Z){if(U==null)throw Error(u(151));for(var ze=null,zt=null,Xe=R,tt=R=0,$n=null,Lt=U.next();Xe!==null&&!Lt.done;tt++,Lt=U.next()){Xe.index>tt?($n=Xe,Xe=null):$n=Xe.sibling;var Fs=A(D,Xe,Lt.value,Z);if(Fs===null){Xe===null&&(Xe=$n);break}e&&Xe&&Fs.alternate===null&&t(D,Xe),R=i(Fs,R,tt),zt===null?ze=Fs:zt.sibling=Fs,zt=Fs,Xe=$n}if(Lt.done)return n(D,Xe),jt&&Jo(D,tt),ze;if(Xe===null){for(;!Lt.done;tt++,Lt=U.next())Lt=J(D,Lt.value,Z),Lt!==null&&(R=i(Lt,R,tt),zt===null?ze=Lt:zt.sibling=Lt,zt=Lt);return jt&&Jo(D,tt),ze}for(Xe=l(Xe);!Lt.done;tt++,Lt=U.next())Lt=q(Xe,D,tt,Lt.value,Z),Lt!==null&&(e&&($n=Lt.alternate,$n!==null&&Xe.delete($n.key===null?tt:$n.key)),R=i(Lt,R,tt),zt===null?ze=Lt:zt.sibling=Lt,zt=Lt);return e&&Xe.forEach(function(Px){return t(D,Px)}),jt&&Jo(D,tt),ze}function mt(D,R,U,Z){if(typeof U=="object"&&U!==null&&U.type===pe&&U.key===null&&U.props.ref===void 0&&(U=U.props.children),typeof U=="object"&&U!==null){switch(U.$$typeof){case ne:e:{for(var ze=U.key;R!==null;){if(R.key===ze){if(ze=U.type,ze===pe){if(R.tag===7){n(D,R.sibling),Z=a(R,U.props.children),Es(Z,U),Z.return=D,D=Z;break e}}else if(R.elementType===ze||typeof ze=="object"&&ze!==null&&ze.$$typeof===Se&&ka(ze)===R.type){n(D,R.sibling),Z=a(R,U.props),Es(Z,U),Z.return=D,D=Z;break e}n(D,R);break}else t(D,R);R=R.sibling}U.type===pe?(Z=fn(U.props.children,D.mode,Z,U.key),Es(Z,U),Z.return=D,D=Z):(Z=il(U.type,U.key,U.props,null,D.mode,Z),Es(Z,U),Z.return=D,D=Z)}return _(D);case Ce:e:{for(ze=U.key;R!==null;){if(R.key===ze)if(R.tag===4&&R.stateNode.containerInfo===U.containerInfo&&R.stateNode.implementation===U.implementation){n(D,R.sibling),Z=a(R,U.children||[]),Z.return=D,D=Z;break e}else{n(D,R);break}else t(D,R);R=R.sibling}Z=zn(U,D.mode,Z),Z.return=D,D=Z}return _(D);case Se:return U=ka(U),mt(D,R,U,Z)}if(mn(U))return ve(D,R,U,Z);if(We(U)){if(ze=We(U),typeof ze!="function")throw Error(u(150));return U=ze.call(U),Le(D,R,U,Z)}if(typeof U.then=="function")return mt(D,R,$c(U),Z);if(U.$$typeof===ue)return mt(D,R,Oc(D,U),Z);Uc(D,U)}return typeof U=="string"&&U!==""||typeof U=="number"||typeof U=="bigint"?(U=""+U,R!==null&&R.tag===6?(n(D,R.sibling),Z=a(R,U),Z.return=D,D=Z):(n(D,R),Z=No(U,D.mode,Z),Z.return=D,D=Z),_(D)):n(D,R)}return function(D,R,U,Z){try{ji=0;var ze=mt(D,R,U,Z);return _r=null,ze}catch(Xe){if(Xe===fr||Xe===Lc)throw Xe;var zt=He(29,Xe,null,D.mode);return zt.lanes=Z,zt.return=D,zt}finally{}}}var Na=mm(!0),hm=mm(!1),Ts=!1;function xd(e){e.updateQueue={baseState:e.memoizedState,firstBaseUpdate:null,lastBaseUpdate:null,shared:{pending:null,lanes:0,hiddenCallbacks:null},callbacks:null}}function bd(e,t){e=e.updateQueue,t.updateQueue===e&&(t.updateQueue={baseState:e.baseState,firstBaseUpdate:e.firstBaseUpdate,lastBaseUpdate:e.lastBaseUpdate,shared:e.shared,callbacks:null})}function Rs(e){return{lane:e,tag:0,payload:null,callback:null,next:null}}function zs(e,t,n){var l=e.updateQueue;if(l===null)return null;if(l=l.shared,(qt&2)!==0){var a=l.pending;return a===null?t.next=t:(t.next=a.next,a.next=t),l.pending=t,t=St(e),st(e,null,n),t}return dt(e,l,t,n),St(e)}function ki(e,t,n){if(t=t.updateQueue,t!==null&&(t=t.shared,(n&4194048)!==0)){var l=t.lanes;l&=e.pendingLanes,n|=l,t.lanes=n,Yn(e,n)}}function vd(e,t){var n=e.updateQueue,l=e.alternate;if(l!==null&&(l=l.updateQueue,n===l)){var a=null,i=null;if(n=n.firstBaseUpdate,n!==null){do{var _={lane:n.lane,tag:n.tag,payload:n.payload,callback:null,next:null};i===null?a=i=_:i=i.next=_,n=n.next}while(n!==null);i===null?a=i=t:i=i.next=t}else a=i=t;n={baseState:l.baseState,firstBaseUpdate:a,lastBaseUpdate:i,shared:l.shared,callbacks:l.callbacks},e.updateQueue=n;return}e=n.lastBaseUpdate,e===null?n.firstBaseUpdate=t:e.next=t,n.lastBaseUpdate=t}var wd=!1;function Ci(){if(wd){var e=dr;if(e!==null)throw e}}function Ni(e,t,n,l){wd=!1;var a=e.updateQueue;Ts=!1;var i=a.firstBaseUpdate,_=a.lastBaseUpdate,y=a.shared.pending;if(y!==null){a.shared.pending=null;var M=y,L=M.next;M.next=null,_===null?i=L:_.next=L,_=M;var X=e.alternate;X!==null&&(X=X.updateQueue,y=X.lastBaseUpdate,y!==_&&(y===null?X.firstBaseUpdate=L:y.next=L,X.lastBaseUpdate=M))}if(i!==null){var J=a.baseState;_=0,X=L=M=null,y=i;do{var A=y.lane&-536870913,q=A!==y.lane;if(q?(Rt&A)===A:(l&A)===A){A!==0&&A===Sa&&(wd=!0),X!==null&&(X=X.next={lane:0,tag:y.tag,payload:y.payload,callback:null,next:null});e:{var ve=e,Le=y;A=t;var mt=n;switch(Le.tag){case 1:if(ve=Le.payload,typeof ve=="function"){J=ve.call(mt,J,A);break e}J=ve;break e;case 3:ve.flags=ve.flags&-65537|128;case 0:if(ve=Le.payload,A=typeof ve=="function"?ve.call(mt,J,A):ve,A==null)break e;J=W({},J,A);break e;case 2:Ts=!0}}A=y.callback,A!==null&&(e.flags|=64,q&&(e.flags|=8192),q=a.callbacks,q===null?a.callbacks=[A]:q.push(A))}else q={lane:A,tag:y.tag,payload:y.payload,callback:y.callback,next:null},X===null?(L=X=q,M=J):X=X.next=q,_|=A;if(y=y.next,y===null){if(y=a.shared.pending,y===null)break;q=y,y=q.next,q.next=null,a.lastBaseUpdate=q,a.shared.pending=null}}while(!0);X===null&&(M=J),a.baseState=M,a.firstBaseUpdate=L,a.lastBaseUpdate=X,i===null&&(a.shared.lanes=0),Us|=_,e.lanes=_,e.memoizedState=J}}function gm(e,t){if(typeof e!="function")throw Error(u(191,e));e.call(t)}function pm(e,t){var n=e.callbacks;if(n!==null)for(e.callbacks=null,e=0;e<n.length;e++)gm(n[e],t)}var As=bn(null),Hc=bn(0);function ym(e,t){e=ss,ye(Hc,e),ye(As,t),ss=e|t.baseLanes}function Sd(){ye(Hc,ss),ye(As,As.current)}function jd(){ss=Hc.current,re(As),re(Hc)}var Pn=bn(null),cl=null;function Os(e){var t=e.alternate;ye(el,el.current&1),ye(Pn,e),cl===null&&(t===null||As.current!==null||t.memoizedState!==null)&&(cl=e)}function kd(e){ye(el,el.current),ye(Pn,e),cl===null&&(cl=e)}function xm(e){e.tag===22?(ye(el,el.current),ye(Pn,e),cl===null&&(cl=e)):Ds()}function Ds(){ye(el,el.current),ye(Pn,Pn.current)}function $l(e){re(Pn),cl===e&&(cl=null),re(el)}var el=bn(0);function Mi(e,t){ye(Pn,Pn.current),ye(el,t)}function Cd(e){re(el),re(Pn),cl===e&&(cl=null)}function Yc(e){for(var t=e;t!==null;){if(t.tag===13){var n=t.memoizedState;if(n!==null&&(n=n.dehydrated,n===null||Wf(n)||Kf(n)))return t}else if(t.tag===19&&t.memoizedProps.revealOrder!=="independent"){if((t.flags&128)!==0)return t}else if(t.child!==null){t.child.return=t,t=t.child;continue}if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return null;t=t.return}t.sibling.return=t.return,t=t.sibling}return null}var ts=0,_t=null,tn=null,Ln=null,qc=!1,mr=!1,Ma=!1,Ic=0,Ei=0,hr=null,by=0;function jn(){throw Error(u(321))}function Nd(e,t){if(t===null)return!1;for(var n=0;n<t.length&&n<e.length;n++)if(!gl(e[n],t[n]))return!1;return!0}function Md(e,t,n,l,a,i){return ts=i,_t=t,t.memoizedState=null,t.updateQueue=null,t.lanes=0,Re.H=e===null||e.memoizedState===null?nh:lh,Ma=!1,i=n(l,a),Ma=!1,mr&&(i=vm(t,n,l,a)),bm(e),i}function bm(e){Re.H=Kc;var t=tn!==null&&tn.next!==null;if(ts=0,Ln=tn=_t=null,qc=!1,Ei=0,hr=null,t)throw Error(u(300));e===null||Bn||(e=e.dependencies,e!==null&&Ac(e)&&(Bn=!0))}function vm(e,t,n,l){_t=e;var a=0;do{if(mr&&(hr=null),Ei=0,mr=!1,25<=a)throw Error(u(301));if(a+=1,Ln=tn=null,e.updateQueue!=null){var i=e.updateQueue;i.lastEffect=null,i.events=null,i.stores=null,i.memoCache!=null&&(i.memoCache.index=0)}Re.H=My,i=t(n,l)}while(mr);return i}function vy(){var e=Re.H,t=e.useState()[0];return t=typeof t.then=="function"?Ti(t):t,e=e.useState()[0],(tn!==null?tn.memoizedState:null)!==e&&(_t.flags|=1024),t}function Ed(){var e=Ic!==0;return Ic=0,e}function Td(e,t,n){t.updateQueue=e.updateQueue,t.flags&=-2053,e.lanes&=~n}function Rd(e){if(qc){for(e=e.memoizedState;e!==null;){var t=e.queue;t!==null&&(t.pending=null),e=e.next}qc=!1}ts=0,Ln=tn=_t=null,mr=!1,Ei=Ic=0,hr=null}function pl(){var e={memoizedState:null,baseState:null,baseQueue:null,queue:null,next:null};return Ln===null?_t.memoizedState=Ln=e:Ln=Ln.next=e,Ln}function An(){if(tn===null){var e=_t.alternate;e=e!==null?e.memoizedState:null}else e=tn.next;var t=Ln===null?_t.memoizedState:Ln.next;if(t!==null)Ln=t,tn=e;else{if(e===null)throw _t.alternate===null?Error(u(467)):Error(u(310));tn=e,e={memoizedState:tn.memoizedState,baseState:tn.baseState,baseQueue:tn.baseQueue,queue:tn.queue,next:null},Ln===null?_t.memoizedState=Ln=e:Ln=Ln.next=e}return Ln}function Xc(){return{lastEffect:null,events:null,stores:null,memoCache:null}}function Ti(e){var t=Ei;return Ei+=1,hr===null&&(hr=[]),e=dm(hr,e,t),t=_t,(Ln===null?t.memoizedState:Ln.next)===null&&(t=t.alternate,Re.H=t===null||t.memoizedState===null?nh:lh),e}function Qc(e){if(e!==null&&typeof e=="object"){if(typeof e.then=="function")return Ti(e);if(e.$$typeof===Xt)return;if(e.$$typeof===ue)return Jn(e)}throw Error(u(438,String(e)))}function zd(e){var t=null,n=_t.updateQueue;if(n!==null&&(t=n.memoCache),t==null){var l=_t.alternate;l!==null&&(l=l.updateQueue,l!==null&&(l=l.memoCache,l!=null&&(t={data:l.data.map(function(a){return a.slice()}),index:0})))}if(t==null&&(t={data:[],index:0}),n===null&&(n=Xc(),_t.updateQueue=n),n.memoCache=t,n=t.data[t.index],n===void 0)for(n=t.data[t.index]=Array(e),l=0;l<e;l++)n[l]=Ye;return t.index++,n}function ns(e,t){return typeof t=="function"?t(e):t}function Gc(e){var t=An();return Ad(t,tn,e)}function Ad(e,t,n){var l=e.queue;if(l===null)throw Error(u(311));l.lastRenderedReducer=n;var a=e.baseQueue,i=l.pending;if(i!==null){if(a!==null){var _=a.next;a.next=i.next,i.next=_}t.baseQueue=a=i,l.pending=null}if(i=e.baseState,a===null)e.memoizedState=i;else{t=a.next;var y=_=null,M=null,L=t,X=!1;do{var J=L.lane&-536870913;if(J!==L.lane?(Rt&J)===J:(ts&J)===J){var A=L.revertLane;if(A===0)M!==null&&(M=M.next={lane:0,revertLane:0,gesture:null,action:L.action,hasEagerState:L.hasEagerState,eagerState:L.eagerState,next:null}),J===Sa&&(X=!0);else if((ts&A)===A){L=L.next,A===Sa&&(X=!0);continue}else J={lane:0,revertLane:L.revertLane,gesture:null,action:L.action,hasEagerState:L.hasEagerState,eagerState:L.eagerState,next:null},M===null?(y=M=J,_=i):M=M.next=J,_t.lanes|=A,Us|=A;J=L.action,Ma&&n(i,J),i=L.hasEagerState?L.eagerState:n(i,J)}else A={lane:J,revertLane:L.revertLane,gesture:L.gesture,action:L.action,hasEagerState:L.hasEagerState,eagerState:L.eagerState,next:null},M===null?(y=M=A,_=i):M=M.next=A,_t.lanes|=J,Us|=J;L=L.next}while(L!==null&&L!==t);if(M===null?_=i:M.next=y,!gl(i,e.memoizedState)&&(Bn=!0,X&&(n=dr,n!==null)))throw n;e.memoizedState=i,e.baseState=_,e.baseQueue=M,l.lastRenderedState=i}return a===null&&(l.lanes=0),[e.memoizedState,l.dispatch]}function Od(e){var t=An(),n=t.queue;if(n===null)throw Error(u(311));n.lastRenderedReducer=e;var l=n.dispatch,a=n.pending,i=t.memoizedState;if(a!==null){n.pending=null;var _=a=a.next;do i=e(i,_.action),_=_.next;while(_!==a);gl(i,t.memoizedState)||(Bn=!0),t.memoizedState=i,t.baseQueue===null&&(t.baseState=i),n.lastRenderedState=i}return[i,l]}function wm(e,t,n){var l=_t,a=An(),i=jt;if(i){if(n===void 0)throw Error(u(407));n=n()}else n=t();var _=!gl((tn||a).memoizedState,n);if(_&&(a.memoizedState=n,Bn=!0),a=a.queue,Bd(km.bind(null,l,a,e),[e]),e=a.getSnapshot!==t||_||Ln!==null&&(Ln.memoizedState.tag&1)!==0,gr(e?9:8,{destroy:void 0},jm.bind(null,l,a,n,t),null),e){if(l.flags|=2048,sn===null)throw Error(u(349));i||(ts&127)!==0||Sm(l,t,n)}return n}function Sm(e,t,n){e.flags|=16384,e={getSnapshot:t,value:n},t=_t.updateQueue,t===null?(t=Xc(),_t.updateQueue=t,t.stores=[e]):(n=t.stores,n===null?t.stores=[e]:n.push(e))}function jm(e,t,n,l){t.value=n,t.getSnapshot=l,Cm(t)&&Nm(e)}function km(e,t,n){return n(function(){Cm(t)&&Nm(e)})}function Cm(e){var t=e.getSnapshot;e=e.value;try{var n=t();return!gl(e,n)}catch{return!0}}function Nm(e){var t=Ve(e,2);t!==null&&Tl(t,e,2)}function Dd(e){var t=pl();if(typeof e=="function"){var n=e;if(e=n(),Ma){je(!0);try{n()}finally{je(!1)}}}return t.memoizedState=t.baseState=e,t.queue={pending:null,lanes:0,dispatch:null,lastRenderedReducer:ns,lastRenderedState:e},t}function Mm(e,t,n,l){return e.baseState=n,Ad(e,tn,typeof l=="function"?l:ns)}function wy(e,t,n,l,a){if(Wc(e))throw Error(u(485));if(e=t.action,e!==null){var i={payload:a,action:e,next:null,isTransition:!0,status:"pending",value:null,reason:null,listeners:[],then:function(_){i.listeners.push(_)}};Re.T!==null?n(!0):i.isTransition=!1,l(i),n=t.pending,n===null?(i.next=t.pending=i,Em(t,i)):(i.next=n.next,t.pending=n.next=i)}}function Em(e,t){var n=t.action,l=t.payload,a=e.state;if(t.isTransition){var i=Re.T,_={};_.types=i!==null?i.types:null,Re.T=_;try{var y=n(a,l),M=Re.S;M!==null&&M(_,y),Tm(e,t,y)}catch(L){Ld(e,t,L)}finally{i!==null&&_.types!==null&&(i.types=_.types),Re.T=i}}else try{i=n(a,l),Tm(e,t,i)}catch(L){Ld(e,t,L)}}function Tm(e,t,n){n!==null&&typeof n=="object"&&typeof n.then=="function"?n.then(function(l){Rm(e,t,l)},function(l){return Ld(e,t,l)}):Rm(e,t,n)}function Rm(e,t,n){t.status="fulfilled",t.value=n,zm(t),e.state=n,t=e.pending,t!==null&&(n=t.next,n===t?e.pending=null:(n=n.next,t.next=n,Em(e,n)))}function Ld(e,t,n){var l=e.pending;if(e.pending=null,l!==null){l=l.next;do t.status="rejected",t.reason=n,zm(t),t=t.next;while(t!==l)}e.action=null}function zm(e){e=e.listeners;for(var t=0;t<e.length;t++)(0,e[t])()}function Am(e,t){return t}function Om(e,t){if(jt){var n=sn.formState;if(n!==null){e:{var l=_t;if(jt){if(rn){t:{for(var a=rn,i=Kl;a.nodeType!==8;){if(!i){a=null;break t}if(a=Jl(a.nextSibling),a===null){a=null;break t}}i=a.data,a=i==="F!"||i==="F"?a:null}if(a){rn=Jl(a.nextSibling),l=a.data==="F!";break e}}Ns(l)}l=!1}l&&(t=n[0])}}return n=pl(),n.memoizedState=n.baseState=t,l={pending:null,lanes:0,dispatch:null,lastRenderedReducer:Am,lastRenderedState:t},n.queue=l,n=Pm.bind(null,_t,l),l.dispatch=n,l=Dd(!1),i=qd.bind(null,_t,!1,l.queue),l=pl(),a={state:t,dispatch:null,action:e,pending:null},l.queue=a,n=wy.bind(null,_t,a,i,n),a.dispatch=n,l.memoizedState=e,[t,n,!1]}function Dm(e){var t=An();return Lm(t,tn,e)}function Lm(e,t,n){if(t=Ad(e,t,Am)[0],e=Gc(ns)[0],typeof t=="object"&&t!==null&&typeof t.then=="function")try{var l=Ti(t)}catch(_){throw _===fr?Lc:_}else l=t;t=An();var a=t.queue,i=a.dispatch;return n!==t.memoizedState&&(_t.flags|=2048,gr(9,{destroy:void 0},Sy.bind(null,a,n),null)),[l,i,e]}function Sy(e,t){e.action=t}function Bm(e){var t=An(),n=tn;if(n!==null)return Lm(t,n,e);An(),t=t.memoizedState,n=An();var l=n.queue.dispatch;return n.memoizedState=e,[t,l,!1]}function gr(e,t,n,l){return e={tag:e,create:n,deps:l,inst:t,next:null},t=_t.updateQueue,t===null&&(t=Xc(),_t.updateQueue=t),n=t.lastEffect,n===null?t.lastEffect=e.next=e:(l=n.next,n.next=e,e.next=l,t.lastEffect=e),e}function $m(){return An().memoizedState}function Vc(e,t,n,l){var a=pl();_t.flags|=e,a.memoizedState=gr(1|t,{destroy:void 0},n,l===void 0?null:l)}function Fc(e,t,n,l){var a=An();l=l===void 0?null:l;var i=a.memoizedState.inst;tn!==null&&l!==null&&Nd(l,tn.memoizedState.deps)?a.memoizedState=gr(t,i,n,l):(_t.flags|=e,a.memoizedState=gr(1|t,i,n,l))}function Um(e,t){Vc(8390656,8,e,t)}function Bd(e,t){Fc(2048,8,e,t)}function jy(e){_t.flags|=4;var t=_t.updateQueue;if(t===null)t=Xc(),_t.updateQueue=t,t.events=[e];else{var n=t.events;n===null?t.events=[e]:n.push(e)}}function Hm(e){var t=An().memoizedState;return jy({ref:t,nextImpl:e}),function(){if((qt&2)!==0)throw Error(u(440));return t.impl.apply(void 0,arguments)}}function Ym(e,t){return Fc(4,2,e,t)}function qm(e,t){return Fc(4,4,e,t)}function Im(e,t){if(typeof t=="function"){e=e();var n=t(e);return function(){typeof n=="function"?n():t(null)}}if(t!=null)return e=e(),t.current=e,function(){t.current=null}}function Xm(e,t,n){n=n!=null?n.concat([e]):null,Fc(4,4,Im.bind(null,t,e),n)}function $d(){}function Qm(e,t){var n=An();t=t===void 0?null:t;var l=n.memoizedState;return t!==null&&Nd(t,l[1])?l[0]:(n.memoizedState=[e,t],e)}function Gm(e,t){var n=An();t=t===void 0?null:t;var l=n.memoizedState;if(t!==null&&Nd(t,l[1]))return l[0];if(l=e(),Ma){je(!0);try{e()}finally{je(!1)}}return n.memoizedState=[l,t],l}function Ud(e,t,n){return n===void 0||(ts&1073741824)!==0&&(Rt&261930)===0?e.memoizedState=t:(e.memoizedState=n,e=ng(),_t.lanes|=e,Us|=e,n)}function Vm(e,t,n,l){return gl(n,t)?n:As.current!==null?(e=Ud(e,n,l),gl(e,t)||(Bn=!0),e):(ts&106)===0||(ts&1073741824)!==0&&(Rt&261930)===0?(Bn=!0,e.memoizedState=n):(e=ng(),_t.lanes|=e,Us|=e,t)}function Fm(e,t,n,l,a){var i=Ue.p;Ue.p=i!==0&&8>i?i:8;var _=Re.T,y={};y.types=_!==null?_.types:null,Re.T=y,qd(e,!1,t,n);try{var M=a(),L=Re.S;if(L!==null&&L(y,M),M!==null&&typeof M=="object"&&typeof M.then=="function"){var X=xy(M,l);Ri(e,t,X,ql(e))}else Ri(e,t,l,ql(e))}catch(J){Ri(e,t,{then:function(){},status:"rejected",reason:J},ql())}finally{Ue.p=i,_!==null&&y.types!==null&&(_.types=y.types),Re.T=_}}function ky(){}function Hd(e,t,n,l){if(e.tag!==5)throw Error(u(476));var a=Wm(e).queue;Fm(e,a,t,un,n===null?ky:function(){return Km(e),n(l)})}function Wm(e){var t=e.memoizedState;if(t!==null)return t;t={memoizedState:un,baseState:un,baseQueue:null,queue:{pending:null,lanes:0,dispatch:null,lastRenderedReducer:ns,lastRenderedState:un},next:null};var n={};return t.next={memoizedState:n,baseState:n,baseQueue:null,queue:{pending:null,lanes:0,dispatch:null,lastRenderedReducer:ns,lastRenderedState:n},next:null},e.memoizedState=t,e=e.alternate,e!==null&&(e.memoizedState=t),t}function Km(e){var t=Wm(e);t.next===null&&(t=e.alternate.memoizedState),Ri(e,t.next.queue,{},ql())}function Yd(){return Jn(Dr)}function Zm(){return An().memoizedState}function Jm(){return An().memoizedState}function Cy(e){for(var t=e.return;t!==null;){switch(t.tag){case 24:case 3:var n=ql();e=Rs(n);var l=zs(t,e,n);l!==null&&(Tl(l,t,n),ki(l,t,n)),t={cache:hd()},e.payload=t;return}t=t.return}}function Ny(e,t,n){var l=ql();n={lane:l,revertLane:0,gesture:null,action:n,hasEagerState:!1,eagerState:null,next:null},Wc(e)?eh(t,n):(n=$e(e,t,n,l),n!==null&&(Tl(n,e,l),th(n,t,l)))}function Pm(e,t,n){var l=ql();Ri(e,t,n,l)}function Ri(e,t,n,l){var a={lane:l,revertLane:0,gesture:null,action:n,hasEagerState:!1,eagerState:null,next:null};if(Wc(e))eh(t,a);else{var i=e.alternate;if(e.lanes===0&&(i===null||i.lanes===0)&&(i=t.lastRenderedReducer,i!==null))try{var _=t.lastRenderedState,y=i(_,n);if(a.hasEagerState=!0,a.eagerState=y,gl(y,_))return dt(e,t,a,0),sn===null&&Pe(),!1}catch{}finally{}if(n=$e(e,t,a,l),n!==null)return Tl(n,e,l),th(n,t,l),!0}return!1}function qd(e,t,n,l){if(l={lane:2,revertLane:Af(),gesture:null,action:l,hasEagerState:!1,eagerState:null,next:null},Wc(e)){if(t)throw Error(u(479))}else t=$e(e,n,l,2),t!==null&&Tl(t,e,2)}function Wc(e){var t=e.alternate;return e===_t||t!==null&&t===_t}function eh(e,t){mr=qc=!0;var n=e.pending;n===null?t.next=t:(t.next=n.next,n.next=t),e.pending=t}function th(e,t,n){if((n&4194048)!==0){var l=t.lanes;l&=e.pendingLanes,n|=l,t.lanes=n,Yn(e,n)}}var Kc={readContext:Jn,use:Qc,useCallback:jn,useContext:jn,useEffect:jn,useImperativeHandle:jn,useLayoutEffect:jn,useInsertionEffect:jn,useMemo:jn,useReducer:jn,useRef:jn,useState:jn,useDebugValue:jn,useDeferredValue:jn,useTransition:jn,useSyncExternalStore:jn,useId:jn,useHostTransitionStatus:jn,useFormState:jn,useActionState:jn,useOptimistic:jn,useMemoCache:jn,useCacheRefresh:jn,useEffectEvent:jn},nh={readContext:Jn,use:Qc,useCallback:function(e,t){return pl().memoizedState=[e,t===void 0?null:t],e},useContext:Jn,useEffect:Um,useImperativeHandle:function(e,t,n){n=n!=null?n.concat([e]):null,Vc(4194308,4,Im.bind(null,t,e),n)},useLayoutEffect:function(e,t){return Vc(4194308,4,e,t)},useInsertionEffect:function(e,t){Vc(4,2,e,t)},useMemo:function(e,t){var n=pl();t=t===void 0?null:t;var l=e();if(Ma){je(!0);try{e()}finally{je(!1)}}return n.memoizedState=[l,t],l},useReducer:function(e,t,n){var l=pl();if(n!==void 0){var a=n(t);if(Ma){je(!0);try{n(t)}finally{je(!1)}}}else a=t;return l.memoizedState=l.baseState=a,e={pending:null,lanes:0,dispatch:null,lastRenderedReducer:e,lastRenderedState:a},l.queue=e,e=e.dispatch=Ny.bind(null,_t,e),[l.memoizedState,e]},useRef:function(e){var t=pl();return e={current:e},t.memoizedState=e},useState:function(e){e=Dd(e);var t=e.queue,n=Pm.bind(null,_t,t);return t.dispatch=n,[e.memoizedState,n]},useDebugValue:$d,useDeferredValue:function(e,t){var n=pl();return Ud(n,e,t)},useTransition:function(){var e=Dd(!1);return e=Fm.bind(null,_t,e.queue,!0,!1),pl().memoizedState=e,[!1,e]},useSyncExternalStore:function(e,t,n){var l=_t,a=pl();if(jt){if(n===void 0)throw Error(u(407));n=n()}else{if(n=t(),sn===null)throw Error(u(349));(Rt&127)!==0||Sm(l,t,n)}a.memoizedState=n;var i={value:n,getSnapshot:t};return a.queue=i,Um(km.bind(null,l,i,e),[e]),l.flags|=2048,gr(9,{destroy:void 0},jm.bind(null,l,i,n,t),null),n},useId:function(){var e=pl(),t=sn.identifierPrefix;if(jt){var n=Eo,l=Wl;n=(l&~(1<<32-Gt(l)-1)).toString(32)+n,t="_"+t+"R_"+n,n=Ic++,0<n&&(t+="H"+n.toString(32)),t+="_"}else n=by++,t="_"+t+"r_"+n.toString(32)+"_";return e.memoizedState=t},useHostTransitionStatus:Yd,useFormState:Om,useActionState:Om,useOptimistic:function(e){var t=pl();t.memoizedState=t.baseState=e;var n={pending:null,lanes:0,dispatch:null,lastRenderedReducer:null,lastRenderedState:null};return t.queue=n,t=qd.bind(null,_t,!0,n),n.dispatch=t,[e,t]},useMemoCache:zd,useCacheRefresh:function(){return pl().memoizedState=Cy.bind(null,_t)},useEffectEvent:function(e){var t=pl(),n={impl:e};return t.memoizedState=n,function(){if((qt&2)!==0)throw Error(u(440));return n.impl.apply(void 0,arguments)}}},lh={readContext:Jn,use:Qc,useCallback:Qm,useContext:Jn,useEffect:Bd,useImperativeHandle:Xm,useInsertionEffect:Ym,useLayoutEffect:qm,useMemo:Gm,useReducer:Gc,useRef:$m,useState:function(){return Gc(ns)},useDebugValue:$d,useDeferredValue:function(e,t){var n=An();return Vm(n,tn.memoizedState,e,t)},useTransition:function(){var e=Gc(ns)[0],t=An().memoizedState;return[typeof e=="boolean"?e:Ti(e),t]},useSyncExternalStore:wm,useId:Zm,useHostTransitionStatus:Yd,useFormState:Dm,useActionState:Dm,useOptimistic:function(e,t){var n=An();return Mm(n,tn,e,t)},useMemoCache:zd,useCacheRefresh:Jm,useEffectEvent:Hm},My={readContext:Jn,use:Qc,useCallback:Qm,useContext:Jn,useEffect:Bd,useImperativeHandle:Xm,useInsertionEffect:Ym,useLayoutEffect:qm,useMemo:Gm,useReducer:Od,useRef:$m,useState:function(){return Od(ns)},useDebugValue:$d,useDeferredValue:function(e,t){var n=An();return tn===null?Ud(n,e,t):Vm(n,tn.memoizedState,e,t)},useTransition:function(){var e=Od(ns)[0],t=An().memoizedState;return[typeof e=="boolean"?e:Ti(e),t]},useSyncExternalStore:wm,useId:Zm,useHostTransitionStatus:Yd,useFormState:Bm,useActionState:Bm,useOptimistic:function(e,t){var n=An();return tn!==null?Mm(n,tn,e,t):(n.baseState=e,[e,n.queue.dispatch])},useMemoCache:zd,useCacheRefresh:Jm,useEffectEvent:Hm};function Id(e,t,n,l){t=e.memoizedState,n=n(l,t),n=n==null?t:W({},t,n),e.memoizedState=n,e.lanes===0&&(e.updateQueue.baseState=n)}var Xd={enqueueSetState:function(e,t,n){e=e._reactInternals;var l=ql(),a=Rs(l);a.payload=t,n!=null&&(a.callback=n),t=zs(e,a,l),t!==null&&(Tl(t,e,l),ki(t,e,l))},enqueueReplaceState:function(e,t,n){e=e._reactInternals;var l=ql(),a=Rs(l);a.tag=1,a.payload=t,n!=null&&(a.callback=n),t=zs(e,a,l),t!==null&&(Tl(t,e,l),ki(t,e,l))},enqueueForceUpdate:function(e,t){e=e._reactInternals;var n=ql(),l=Rs(n);l.tag=2,t!=null&&(l.callback=t),t=zs(e,l,n),t!==null&&(Tl(t,e,n),ki(t,e,n))}};function oh(e,t,n,l,a,i,_){return e=e.stateNode,typeof e.shouldComponentUpdate=="function"?e.shouldComponentUpdate(l,i,_):t.prototype&&t.prototype.isPureReactComponent?!ro(n,l)||!ro(a,i):!0}function sh(e,t,n,l){e=t.state,typeof t.componentWillReceiveProps=="function"&&t.componentWillReceiveProps(n,l),typeof t.UNSAFE_componentWillReceiveProps=="function"&&t.UNSAFE_componentWillReceiveProps(n,l),t.state!==e&&Xd.enqueueReplaceState(t,t.state,null)}function Ea(e,t){var n=t;if("ref"in t){n={};for(var l in t)l!=="ref"&&(n[l]=t[l])}if(e=e.defaultProps){n===t&&(n=W({},n));for(var a in e)n[a]===void 0&&(n[a]=e[a])}return n}function ah(e){me(e)}function rh(e){console.error(e)}function ih(e){me(e)}function Zc(e,t){try{var n=e.onUncaughtError;n(t.value,{componentStack:t.stack})}catch(l){setTimeout(function(){throw l})}}function ch(e,t,n){try{var l=e.onCaughtError;l(n.value,{componentStack:n.stack,errorBoundary:t.tag===1?t.stateNode:null})}catch(a){setTimeout(function(){throw a})}}function Qd(e,t,n){return n=Rs(n),n.tag=3,n.payload={element:null},n.callback=function(){Zc(e,t)},n}function uh(e){return e=Rs(e),e.tag=3,e}function dh(e,t,n,l){var a=n.type.getDerivedStateFromError;if(typeof a=="function"){var i=l.value;e.payload=function(){return a(i)},e.callback=function(){ch(t,n,l)}}var _=n.stateNode;_!==null&&typeof _.componentDidCatch=="function"&&(e.callback=function(){ch(t,n,l),typeof a!="function"&&(Hs===null?Hs=new Set([this]):Hs.add(this));var y=l.stack;this.componentDidCatch(l.value,{componentStack:y!==null?y:""})})}function Ey(e,t,n,l,a){if(n.flags|=32768,l!==null&&typeof l=="object"&&typeof l.then=="function"){if(t=n.alternate,t!==null&&va(t,n,a,!0),n=Pn.current,n!==null){switch(n.tag){case 31:case 13:case 19:return cl===null?yu():n.alternate===null&&kn===0&&(kn=3),n.flags&=-257,n.flags|=65536,n.lanes=a,l===Bc?n.flags|=16384:(t=n.updateQueue,t===null?n.updateQueue=new Set([l]):t.add(l),Tf(e,l,a)),!1;case 22:return n.flags|=65536,l===Bc?n.flags|=16384:(t=n.updateQueue,t===null?(t={transitions:null,markerInstances:null,retryQueue:new Set([l])},n.updateQueue=t):(n=t.retryQueue,n===null?t.retryQueue=new Set([l]):n.add(l)),Tf(e,l,a)),!1}throw Error(u(435,n.tag))}return Tf(e,l,a),yu(),!1}if(jt)return t=Pn.current,t!==null?((t.flags&65536)===0&&(t.flags|=256),t.flags|=65536,t.lanes=a,l!==dd&&(e=Error(u(422),{cause:l}),bi(Zn(e,n)))):(l!==dd&&(t=Error(u(423),{cause:l}),bi(Zn(t,n))),e=e.current.alternate,e.flags|=65536,a&=-a,e.lanes|=a,l=Zn(l,n),a=Qd(e.stateNode,l,a),vd(e,a),kn!==4&&(kn=2)),!1;var i=Error(u(520),{cause:l});if(i=Zn(i,n),Ui===null?Ui=[i]:Ui.push(i),kn!==4&&(kn=2),t===null)return!0;l=Zn(l,n),n=t;do{switch(n.tag){case 3:return n.flags|=65536,e=a&-a,n.lanes|=e,e=Qd(n.stateNode,l,e),vd(n,e),!1;case 1:if(t=n.type,i=n.stateNode,(n.flags&128)===0&&(typeof t.getDerivedStateFromError=="function"||i!==null&&typeof i.componentDidCatch=="function"&&(Hs===null||!Hs.has(i))))return n.flags|=65536,a&=-a,n.lanes|=a,a=uh(a),dh(a,e,n,l),vd(n,a),!1;break;case 22:if(n.memoizedState!==null)return n.flags|=65536,!1}n=n.return}while(n!==null);return!1}var Gd=Error(u(461)),Bn=!1;function In(e,t,n,l){t.child=e===null?hm(t,null,n,l):Na(t,e.child,n,l)}function fh(e,t,n,l,a){n=n.render;var i=t.ref;if("ref"in l){var _={};for(var y in l)y!=="ref"&&(_[y]=l[y])}else _=l;return wa(t),l=Md(e,t,n,_,i,a),y=Ed(),e!==null&&!Bn?(Td(e,t,a),ls(e,t,a)):(jt&&y&&Tc(t),t.flags|=1,In(e,t,l,a),t.child)}function _h(e,t,n,l,a){if(e===null){var i=n.type;return typeof i=="function"&&!et(i)&&i.defaultProps===void 0&&n.compare===null?(t.tag=15,t.type=i,mh(e,t,i,l,a)):(e=il(n.type,null,l,t,t.mode,a),e.ref=t.ref,e.return=t,t.child=e)}if(i=e.child,!ef(e,a)){var _=i.memoizedProps;if(n=n.compare,n=n!==null?n:ro,n(_,l)&&e.ref===t.ref)return ls(e,t,a)}return t.flags|=1,e=Ct(i,l),e.ref=t.ref,e.return=t,t.child=e}function mh(e,t,n,l,a){if(e!==null){var i=e.memoizedProps;if(ro(i,l)&&e.ref===t.ref)if(Bn=!1,t.pendingProps=l=i,ef(e,a))(e.flags&131072)!==0&&(Bn=!0);else return t.lanes=e.lanes,ls(e,t,a)}return Vd(e,t,n,l,a)}function hh(e,t,n,l){var a=l.children,i=e!==null?e.memoizedState:null;if(e===null&&t.stateNode===null&&(t.stateNode={_visibility:1,_pendingMarkers:null,_retryCache:null,_transitions:null}),l.mode==="hidden"){if((t.flags&128)!==0){if(i=i!==null?i.baseLanes|n:n,e!==null){for(l=t.child=e.child,a=0;l!==null;)a=a|l.lanes|l.childLanes,l=l.sibling;l=a&~i}else l=0,t.child=null;return gh(e,t,i,n,l)}if((n&536870912)!==0)t.memoizedState={baseLanes:0,cachePool:null},e!==null&&Dc(t,i!==null?i.cachePool:null),i!==null?ym(t,i):Sd(),xm(t);else return l=t.lanes=536870912,gh(e,t,i!==null?i.baseLanes|n:n,n,l)}else i!==null?(Dc(t,i.cachePool),ym(t,i),Ds(),t.memoizedState=null):(e!==null&&Dc(t,null),Sd(),Ds());return In(e,t,a,n),t.child}function zi(e,t){return e!==null&&e.tag===22||t.stateNode!==null||(t.stateNode={_visibility:1,_pendingMarkers:null,_retryCache:null,_transitions:null}),t.sibling}function gh(e,t,n,l,a){var i=pd();return i=i===null?null:{parent:Dn._currentValue,pool:i},t.memoizedState={baseLanes:n,cachePool:i},e!==null&&Dc(t,null),Sd(),xm(t),e!==null&&va(e,t,l,!0),t.childLanes=a,null}function Jc(e,t){return t=Pc({mode:t.mode,children:t.children},e.mode),t.ref=e.ref,e.child=t,t.return=e,t}function ph(e,t,n){return Na(t,e.child,null,n),e=Jc(t,t.pendingProps),e.flags|=2,$l(t),t.memoizedState=null,e}function Ty(e,t,n){var l=t.pendingProps,a=(t.flags&128)!==0;if(t.flags&=-129,e===null){if(jt){if(l.mode==="hidden")return e=Jc(t,l),t.lanes=536870912,e.memoizedState={baseLanes:0,cachePool:null},zi(null,e);if(kd(t),(e=rn)?(e=Ig(e,Kl),e=e!==null&&e.data==="&"?e:null,e!==null&&(t.memoizedState={dehydrated:e,treeContext:Mo!==null?{id:Wl,overflow:Eo}:null,retryLane:536870912,hydrationErrors:null},n=Zo(e),n.return=t,t.child=n,Qn=t,rn=null)):e=null,e===null)throw Ns(t);return t.lanes=536870912,null}return Jc(t,l)}var i=e.memoizedState;if(i!==null){var _=i.dehydrated;if(kd(t),a)if(t.flags&256)t.flags&=-257,t=ph(e,t,n);else if(t.memoizedState!==null)t.child=e.child,t.flags|=128,t=null;else throw Error(u(558));else if(Bn||va(e,t,n,!1),a=(n&e.childLanes)!==0,Bn||a){if(As.current===null){if(l=sn,l!==null&&(_=go(l,n),_!==0&&_!==i.retryLane))throw i.retryLane=_,Ve(e,_),Tl(l,e,_),Gd;yu()}t=ph(e,t,n)}else e=i.treeContext,rn=Jl(_.nextSibling),Qn=t,jt=!0,Cs=null,Kl=!1,e!==null&&lm(t,e),t=Jc(t,l),t.flags|=134221824;return t}return e=Ct(e.child,{mode:l.mode,children:l.children}),e.ref=t.ref,t.child=e,e.return=t,e}function pr(e,t){var n=t.ref;if(n===null)e!==null&&e.ref!==null&&(t.flags|=4194816);else{if(typeof n!="function"&&typeof n!="object")throw Error(u(284));(e===null||e.ref!==n)&&(t.flags|=4194816)}}function Vd(e,t,n,l,a){return wa(t),n=Md(e,t,n,l,void 0,a),l=Ed(),e!==null&&!Bn?(Td(e,t,a),ls(e,t,a)):(jt&&l&&Tc(t),t.flags|=1,In(e,t,n,a),t.child)}function yh(e,t,n,l,a,i){return wa(t),t.updateQueue=null,n=vm(t,l,n,a),bm(e),l=Ed(),e!==null&&!Bn?(Td(e,t,i),ls(e,t,i)):(jt&&l&&Tc(t),t.flags|=1,In(e,t,n,i),t.child)}function xh(e,t,n,l,a){if(wa(t),t.stateNode===null){var i=$t,_=n.contextType;typeof _=="object"&&_!==null&&(i=Jn(_)),i=new n(l,i),t.memoizedState=i.state!==null&&i.state!==void 0?i.state:null,i.updater=Xd,t.stateNode=i,i._reactInternals=t,i=t.stateNode,i.props=l,i.state=t.memoizedState,i.refs={},xd(t),_=n.contextType,i.context=typeof _=="object"&&_!==null?Jn(_):$t,i.state=t.memoizedState,_=n.getDerivedStateFromProps,typeof _=="function"&&(Id(t,n,_,l),i.state=t.memoizedState),typeof n.getDerivedStateFromProps=="function"||typeof i.getSnapshotBeforeUpdate=="function"||typeof i.UNSAFE_componentWillMount!="function"&&typeof i.componentWillMount!="function"||(_=i.state,typeof i.componentWillMount=="function"&&i.componentWillMount(),typeof i.UNSAFE_componentWillMount=="function"&&i.UNSAFE_componentWillMount(),_!==i.state&&Xd.enqueueReplaceState(i,i.state,null),Ni(t,l,i,a),Ci(),i.state=t.memoizedState),typeof i.componentDidMount=="function"&&(t.flags|=4194308),l=!0}else if(e===null){i=t.stateNode;var y=t.memoizedProps,M=Ea(n,y);i.props=M;var L=i.context,X=n.contextType;_=$t,typeof X=="object"&&X!==null&&(_=Jn(X));var J=n.getDerivedStateFromProps;X=typeof J=="function"||typeof i.getSnapshotBeforeUpdate=="function",y=t.pendingProps!==y,X||typeof i.UNSAFE_componentWillReceiveProps!="function"&&typeof i.componentWillReceiveProps!="function"||(y||L!==_)&&sh(t,i,l,_),Ts=!1;var A=t.memoizedState;i.state=A,Ni(t,l,i,a),Ci(),L=t.memoizedState,y||A!==L||Ts?(typeof J=="function"&&(Id(t,n,J,l),L=t.memoizedState),(M=Ts||oh(t,n,M,l,A,L,_))?(X||typeof i.UNSAFE_componentWillMount!="function"&&typeof i.componentWillMount!="function"||(typeof i.componentWillMount=="function"&&i.componentWillMount(),typeof i.UNSAFE_componentWillMount=="function"&&i.UNSAFE_componentWillMount()),typeof i.componentDidMount=="function"&&(t.flags|=4194308)):(typeof i.componentDidMount=="function"&&(t.flags|=4194308),t.memoizedProps=l,t.memoizedState=L),i.props=l,i.state=L,i.context=_,l=M):(typeof i.componentDidMount=="function"&&(t.flags|=4194308),l=!1)}else{i=t.stateNode,bd(e,t),_=t.memoizedProps,X=Ea(n,_),i.props=X,J=t.pendingProps,A=i.context,L=n.contextType,M=$t,typeof L=="object"&&L!==null&&(M=Jn(L)),y=n.getDerivedStateFromProps,(L=typeof y=="function"||typeof i.getSnapshotBeforeUpdate=="function")||typeof i.UNSAFE_componentWillReceiveProps!="function"&&typeof i.componentWillReceiveProps!="function"||(_!==J||A!==M)&&sh(t,i,l,M),Ts=!1,A=t.memoizedState,i.state=A,Ni(t,l,i,a),Ci();var q=t.memoizedState;_!==J||A!==q||Ts||e!==null&&e.dependencies!==null&&Ac(e.dependencies)?(typeof y=="function"&&(Id(t,n,y,l),q=t.memoizedState),(X=Ts||oh(t,n,X,l,A,q,M)||e!==null&&e.dependencies!==null&&Ac(e.dependencies))?(L||typeof i.UNSAFE_componentWillUpdate!="function"&&typeof i.componentWillUpdate!="function"||(typeof i.componentWillUpdate=="function"&&i.componentWillUpdate(l,q,M),typeof i.UNSAFE_componentWillUpdate=="function"&&i.UNSAFE_componentWillUpdate(l,q,M)),typeof i.componentDidUpdate=="function"&&(t.flags|=4),typeof i.getSnapshotBeforeUpdate=="function"&&(t.flags|=1024)):(typeof i.componentDidUpdate!="function"||_===e.memoizedProps&&A===e.memoizedState||(t.flags|=4),typeof i.getSnapshotBeforeUpdate!="function"||_===e.memoizedProps&&A===e.memoizedState||(t.flags|=1024),t.memoizedProps=l,t.memoizedState=q),i.props=l,i.state=q,i.context=M,l=X):(typeof i.componentDidUpdate!="function"||_===e.memoizedProps&&A===e.memoizedState||(t.flags|=4),typeof i.getSnapshotBeforeUpdate!="function"||_===e.memoizedProps&&A===e.memoizedState||(t.flags|=1024),l=!1)}return i=l,pr(e,t),l=(t.flags&128)!==0,i||l?(i=t.stateNode,n=l&&typeof n.getDerivedStateFromError!="function"?null:i.render(),t.flags|=1,e!==null&&l?(t.child=Na(t,e.child,null,a),t.child=Na(t,null,n,a)):In(e,t,n,a),t.memoizedState=i.state,e=t.child):e=ls(e,t,a),e}function bh(e,t,n,l){return xa(),t.flags|=256,In(e,t,n,l),t.child}var Fd={dehydrated:null,treeContext:null,retryLane:0,hydrationErrors:null};function Wd(e){return{baseLanes:e,cachePool:cm()}}function Kd(e,t,n){return e=e!==null?e.childLanes&~n:0,t&&(e|=Yl),e}function vh(e,t,n){var l=t.pendingProps,a=!1,i=(t.flags&128)!==0,_;if((_=i)||(_=e!==null&&e.memoizedState===null?!1:(el.current&2)!==0),_&&(a=!0,t.flags&=-129),_=(t.flags&32)!==0,t.flags&=-33,e===null){if(jt){if(a?Os(t):Ds(),(e=rn)?(e=Ig(e,Kl),e=e!==null&&e.data!=="&"?e:null,e!==null&&(t.memoizedState={dehydrated:e,treeContext:Mo!==null?{id:Wl,overflow:Eo}:null,retryLane:536870912,hydrationErrors:null},n=Zo(e),n.return=t,t.child=n,Qn=t,rn=null)):e=null,e===null)throw Ns(t);return Kf(e)?t.lanes=32:t.lanes=536870912,null}return i=l.children,l=l.fallback,a?(Ds(),a=t.mode,i=Pc({mode:"hidden",children:i},a),l=fn(l,a,n,null),i.return=t,l.return=t,i.sibling=l,t.child=i,l=t.child,l.memoizedState=Wd(n),l.childLanes=Kd(e,_,n),t.memoizedState=Fd,zi(null,l)):(Os(t),Zd(t,i))}var y=e.memoizedState;if(y!==null){var M=y.dehydrated;if(M!==null)return Ry(e,t,i,_,l,M,y,n)}return a?(Ds(),a=l.fallback,i=t.mode,y=e.child,M=y.sibling,l=Ct(y,{mode:"hidden",children:l.children}),l.subtreeFlags=y.subtreeFlags&1206910976,M!==null?a=Ct(M,a):(a=fn(a,i,n,null),a.flags|=2),a.return=t,l.return=t,l.sibling=a,t.child=l,zi(null,l),l=t.child,a=e.child.memoizedState,a===null?a=Wd(n):(i=a.cachePool,i!==null?(y=Dn._currentValue,i=i.parent!==y?{parent:y,pool:y}:i):i=cm(),a={baseLanes:a.baseLanes|n,cachePool:i}),l.memoizedState=a,l.childLanes=Kd(e,_,n),t.memoizedState=Fd,zi(e.child,l)):(Os(t),n=e.child,e=n.sibling,n=Ct(n,{mode:"visible",children:l.children}),n.return=t,n.sibling=null,e!==null&&(_=t.deletions,_===null?(t.deletions=[e],t.flags|=16):_.push(e)),t.child=n,t.memoizedState=null,n)}function Zd(e,t){return t=Pc({mode:"visible",children:t},e.mode),t.return=e,e.child=t}function Pc(e,t){return e=He(22,e,null,t),e.lanes=0,e}function eu(e,t,n){return Na(t,e.child,null,n),e=Zd(t,t.pendingProps.children),e.flags|=2,t.memoizedState=null,e}function Ry(e,t,n,l,a,i,_,y){if(n)return t.flags&256?(Os(t),t.flags&=-257,eu(e,t,y)):t.memoizedState!==null?(Ds(),t.child=e.child,t.flags|=128,null):(Ds(),i=a.fallback,_=t.mode,a=Pc({mode:"visible",children:a.children},_),i=fn(i,_,y,null),i.flags|=2,a.return=t,i.return=t,a.sibling=i,t.child=a,Na(t,e.child,null,y),a=t.child,a.memoizedState=Wd(y),a.childLanes=Kd(e,l,y),t.memoizedState=Fd,zi(null,a));if(Os(t),Kf(i)){if(l=i.nextSibling&&i.nextSibling.dataset,l)var M=l.dgst;return l=M,l!==""&&(a=Error(u(419)),a.stack="",a.digest=l,bi({value:a,source:null,stack:null})),eu(e,t,y)}if(Bn||va(e,t,y,!1),l=(y&e.childLanes)!==0,Bn||l){if(As.current!==null)return eu(e,t,y);if(l=sn,l!==null&&(a=go(l,y),a!==0&&a!==_.retryLane))throw _.retryLane=a,Ve(e,a),Tl(l,e,a),Gd;return Wf(i)||yu(),eu(e,t,y)}return Wf(i)?(t.flags|=192,t.child=e.child,null):(e=_.treeContext,rn=Jl(i.nextSibling),Qn=t,jt=!0,Cs=null,Kl=!1,e!==null&&lm(t,e),t=Zd(t,a.children),t.flags|=134221824,t)}function wh(e,t,n){e.lanes|=t;var l=e.alternate;l!==null&&(l.lanes|=t),zc(e.return,t,n)}function Sh(e){for(var t=null;e!==null;){var n=e.alternate;n!==null&&Yc(n)===null&&(t=e),e=e.sibling}return t}function tu(e,t,n,l,a,i){var _=e.memoizedState;_===null?e.memoizedState={isBackwards:t,rendering:null,renderingStartTime:0,last:l,tail:n,tailMode:a,treeForkCount:i}:(_.isBackwards=t,_.rendering=null,_.renderingStartTime=0,_.last=l,_.tail=n,_.tailMode=a,_.treeForkCount=i)}function Jd(e){var t=e.child;for(e.child=null;t!==null;){var n=t.sibling;t.sibling=e.child,e.child=t,t=n}}function Pd(e,t,n){var l=t.pendingProps,a=l.revealOrder,i=l.tail;l=l.children;var _=el.current;if(t.flags&128)return Mi(t,_),null;var y=(_&2)!==0;if(y?(_=_&1|2,t.flags|=128):_&=1,Mi(t,_),a==="backwards"&&e!==null?(Jd(e),In(e,t,l,n),Jd(e)):In(e,t,l,n),l=jt?Bl:0,!y&&e!==null&&(e.flags&128)!==0)e:for(e=t.child;e!==null;){if(e.tag===13)e.memoizedState!==null&&wh(e,n,t);else if(e.tag===19)wh(e,n,t);else if(e.child!==null){e.child.return=e,e=e.child;continue}if(e===t)break e;for(;e.sibling===null;){if(e.return===null||e.return===t)break e;e=e.return}e.sibling.return=e.return,e=e.sibling}switch(a){case"backwards":n=Sh(t.child),n===null?(a=t.child,t.child=null):(a=n.sibling,n.sibling=null,Jd(t)),tu(t,!0,a,null,i,l);break;case"unstable_legacy-backwards":for(n=null,a=t.child,t.child=null;a!==null;){if(e=a.alternate,e!==null&&Yc(e)===null){t.child=a;break}e=a.sibling,a.sibling=n,n=a,a=e}tu(t,!0,n,null,i,l);break;case"together":tu(t,!1,null,null,void 0,l);break;case"independent":t.memoizedState=null;break;default:n=Sh(t.child),n===null?(a=t.child,t.child=null):(a=n.sibling,n.sibling=null),tu(t,!1,a,n,i,l)}return t.child}function jh(e,t,n){var l=t.pendingProps;return Ms(t,t.type,l.value),In(e,t,l.children,n),t.child}function ls(e,t,n){if(e!==null&&(t.dependencies=e.dependencies),Us|=t.lanes,(n&t.childLanes)===0)if(e!==null){if(va(e,t,n,!1),(n&t.childLanes)===0)return null}else return null;if(e!==null&&t.child!==e.child)throw Error(u(153));if(t.child!==null){for(e=t.child,n=Ct(e,e.pendingProps),t.child=n,n.return=t;e.sibling!==null;)e=e.sibling,n=n.sibling=Ct(e,e.pendingProps),n.return=t;n.sibling=null}return t.child}function ef(e,t){return(e.lanes&t)!==0?!0:(e=e.dependencies,!!(e!==null&&Ac(e)))}function zy(e,t,n){switch(t.tag){case 3:it(t,t.stateNode.containerInfo),Ms(t,Dn,e.memoizedState.cache),xa();break;case 27:case 5:Nt(t);break;case 4:it(t,t.stateNode.containerInfo);break;case 10:Ms(t,t.type,t.memoizedProps.value);break;case 31:if(t.memoizedState!==null)return t.flags|=128,kd(t),null;break;case 13:var l=t.memoizedState;if(l!==null){if(l.dehydrated!==null)return Os(t),t.flags|=128,null;l=va(e,t,n,!1);var a=t.child.childLanes;return l||(n&a)!==0?vh(e,t,n):(Os(t),e=ls(e,t,n),e!==null?e.sibling:null)}Os(t);break;case 19:if(t.flags&128)return Pd(e,t,n);if(a=(e.flags&128)!==0,l=(n&t.childLanes)!==0,l||(va(e,t,n,!1),l=(n&t.childLanes)!==0),a){if(l)return Pd(e,t,n);t.flags|=128}if(a=t.memoizedState,a!==null&&(a.rendering=null,a.tail=null,a.lastEffect=null),Mi(t,el.current),l)break;return null;case 22:return t.lanes=0,hh(e,t,n,t.pendingProps);case 24:Ms(t,Dn,e.memoizedState.cache)}return ls(e,t,n)}function kh(e,t,n){if(e!==null)if(e.memoizedProps!==t.pendingProps)Bn=!0;else{if(!ef(e,n)&&(t.flags&128)===0)return Bn=!1,zy(e,t,n);Bn=(e.flags&131072)!==0}else Bn=!1,jt&&(t.flags&1048576)!==0&&nm(t,Bl,t.index);switch(t.lanes=0,t.tag){case 16:e:{var l=t.pendingProps;if(e=ka(t.elementType),t.type=e,typeof e=="function")et(e)?(l=Ea(e,l),t.tag=1,t=xh(null,t,e,l,n)):(t.tag=0,t=Vd(null,t,e,l,n));else{if(e!=null){var a=e.$$typeof;if(a===G){t.tag=11,t=fh(null,t,e,l,n);break e}else if(a===Oe){t.tag=14,t=_h(null,t,e,l,n);break e}else if(a===ue){t.tag=10,t.type=e,t=jh(null,t,n);break e}}throw t=Qt(e)||e,Error(u(306,t,""))}}return t;case 0:return Vd(e,t,t.type,t.pendingProps,n);case 1:return l=t.type,a=Ea(l,t.pendingProps),xh(e,t,l,a,n);case 3:e:{if(it(t,t.stateNode.containerInfo),e===null)throw Error(u(387));l=t.pendingProps;var i=t.memoizedState;a=i.element,bd(e,t),Ni(t,l,null,n);var _=t.memoizedState;if(l=_.cache,Ms(t,Dn,l),l!==i.cache&&md(t,[Dn],n,!0),Ci(),l=_.element,i.isDehydrated)if(i={element:l,isDehydrated:!1,cache:_.cache},t.updateQueue.baseState=i,t.memoizedState=i,t.flags&256){t=bh(e,t,l,n);break e}else if(l!==a){a=Zn(Error(u(424)),t),bi(a),t=bh(e,t,l,n);break e}else{switch(e=t.stateNode.containerInfo,e.nodeType){case 9:e=e.body;break;default:e=e.nodeName==="HTML"?e.ownerDocument.body:e}for(rn=Jl(e.firstChild),Qn=t,jt=!0,Cs=null,Kl=!0,n=hm(t,null,l,n),t.child=n;n;)n.flags=n.flags&-3|134221824,n=n.sibling}else{if(xa(),l===a){t=ls(e,t,n);break e}In(e,t,l,n)}t=t.child}return t;case 26:return pr(e,t),e===null?(n=Kg(t.type,null,t.pendingProps,null))?t.memoizedState=n:jt||(t.stateNode=Eg(t.type,t.pendingProps,lt.current,t)):t.memoizedState=Kg(t.type,e.memoizedProps,t.pendingProps,e.memoizedState),null;case 27:return Nt(t),e===null&&jt&&(l=t.stateNode=Gg(t.type,t.pendingProps,lt.current),Qn=t,Kl=!0,a=rn,Is(t.type)?(Zf=a,rn=Jl(l.firstChild)):rn=a),In(e,t,t.pendingProps.children,n),pr(e,t),e===null&&(t.flags|=4194304),t.child;case 5:return e===null&&jt&&((a=l=rn)&&(l=Cx(l,t.type,t.pendingProps,Kl),l!==null?(t.stateNode=l,Qn=t,rn=Jl(l.firstChild),Kl=!1,a=!0):a=!1),a||Ns(t)),Nt(t),a=t.type,i=t.pendingProps,_=e!==null?e.memoizedProps:null,l=i.children,qf(a,i)?l=null:_!==null&&qf(a,_)&&(t.flags|=32),t.memoizedState!==null&&(a=Md(e,t,vy,null,null,n),Dr._currentValue=a),pr(e,t),In(e,t,l,n),t.child;case 6:return e===null&&jt&&((e=n=rn)&&(n=Nx(n,t.pendingProps,Kl),n!==null?(t.stateNode=n,Qn=t,rn=null,e=!0):e=!1),e||Ns(t)),null;case 13:return vh(e,t,n);case 4:return it(t,t.stateNode.containerInfo),l=t.pendingProps,e===null?t.child=Na(t,null,l,n):In(e,t,l,n),t.child;case 11:return fh(e,t,t.type,t.pendingProps,n);case 7:return l=t.pendingProps,pr(e,t),In(e,t,l,n),t.child;case 8:return In(e,t,t.pendingProps.children,n),t.child;case 12:return In(e,t,t.pendingProps.children,n),t.child;case 10:return jh(e,t,n);case 9:return a=t.type._context,l=t.pendingProps.children,wa(t),a=Jn(a),l=l(a),t.flags|=1,In(e,t,l,n),t.child;case 14:return _h(e,t,t.type,t.pendingProps,n);case 15:return mh(e,t,t.type,t.pendingProps,n);case 19:return Pd(e,t,n);case 31:return Ty(e,t,n);case 22:return hh(e,t,n,t.pendingProps);case 24:return wa(t),l=Jn(Dn),e===null?(a=pd(),a===null&&(a=sn,i=hd(),a.pooledCache=i,i.refCount++,i!==null&&(a.pooledCacheLanes|=n),a=i),t.memoizedState={parent:l,cache:a},xd(t),Ms(t,Dn,a)):((e.lanes&n)!==0&&(bd(e,t),Ni(t,null,null,n),Ci()),a=e.memoizedState,i=t.memoizedState,a.parent!==l?(a={parent:l,cache:l},t.memoizedState=a,t.lanes===0&&(t.memoizedState=t.updateQueue.baseState=a),Ms(t,Dn,l)):(l=i.cache,Ms(t,Dn,l),l!==a.cache&&md(t,[Dn],n,!0))),In(e,t,t.pendingProps.children,n),t.child;case 30:return t.stateNode===null&&(t.stateNode={autoName:null,paired:null,clones:null,ref:null}),l=t.pendingProps,l.name!=null&&l.name!=="auto"?t.flags|=e===null?18882560:18874368:jt&&Tc(t),e!==null&&e.memoizedProps.name!==l.name?t.flags|=4194816:pr(e,t),In(e,t,l.children,n),t.child;case 29:throw t.pendingProps}throw Error(u(156,t.tag))}function os(e){e.flags|=4}function tf(e,t,n,l,a){var i;if((i=(e.mode&32)!==0)&&(i=n===null?e0(t,l):e0(t,l)&&(l.src!==n.src||l.srcSet!==n.srcSet)),i){if(e.flags|=16777216,(a&335544128)===a)if(e.stateNode.complete)e.flags|=8192;else if(ag())e.flags|=8192;else throw Ca=Bc,yd}else e.flags&=-16777217}function Ch(e,t){if(t.type!=="stylesheet"||(t.state.loading&4)!==0)e.flags&=-16777217;else if(e.flags|=16777216,!t0(t))if(ag())e.flags|=8192;else throw Ca=Bc,yd}function nu(e,t){t!==null&&(e.flags|=4),e.flags&16384&&(t=e.tag!==22?Vr():536870912,e.lanes|=t,wr|=t)}function Ai(e,t){if(!jt)switch(e.tailMode){case"visible":break;case"collapsed":for(var n=e.tail,l=null;n!==null;)n.alternate!==null&&(l=n),n=n.sibling;l===null?t||e.tail===null?e.tail=null:e.tail.sibling=null:l.sibling=null;break;default:for(t=e.tail,n=null;t!==null;)t.alternate!==null&&(n=t),t=t.sibling;n===null?e.tail=null:n.sibling=null}}function cn(e){var t=e.alternate!==null&&e.alternate.child===e.child,n=0,l=0;if(t)for(var a=e.child;a!==null;)n|=a.lanes|a.childLanes,l|=a.subtreeFlags&1206910976,l|=a.flags&1206910976,a.return=e,a=a.sibling;else for(a=e.child;a!==null;)n|=a.lanes|a.childLanes,l|=a.subtreeFlags,l|=a.flags,a.return=e,a=a.sibling;return e.subtreeFlags|=l,e.childLanes=n,t}function Ay(e,t,n){var l=t.pendingProps;switch(ud(t),t.tag){case 16:case 15:case 0:case 11:case 7:case 8:case 12:case 9:case 14:return cn(t),null;case 1:return cn(t),null;case 3:return n=t.stateNode,l=null,e!==null&&(l=e.memoizedState.cache),t.memoizedState.cache!==l&&(t.flags|=2048),es(Dn),Fe(),n.pendingContext&&(n.context=n.pendingContext,n.pendingContext=null),(e===null||e.child===null)&&(ur(t)?os(t):e===null||e.memoizedState.isDehydrated&&(t.flags&256)===0||(t.flags|=1024,fd())),cn(t),null;case 26:var a=t.type,i=t.memoizedState;return e===null?(os(t),i!==null?(cn(t),Ch(t,i)):(cn(t),tf(t,a,null,l,n))):i?i!==e.memoizedState?(os(t),cn(t),Ch(t,i)):(cn(t),t.flags&=-16777217):(e=e.memoizedProps,e!==l&&os(t),cn(t),tf(t,a,e,l,n)),null;case 27:if(an(t),n=lt.current,a=t.type,e!==null&&t.stateNode!=null)e.memoizedProps!==l&&os(t);else{if(!l){if(t.stateNode===null)throw Error(u(166));return cn(t),t.subtreeFlags&=-33554433,null}e=Ge.current,ur(t)?om(t):(e=Gg(a,l,n),t.stateNode=e,os(t))}return cn(t),t.subtreeFlags&=-33554433,null;case 5:if(an(t),a=t.type,e!==null&&t.stateNode!=null)e.memoizedProps!==l&&os(t);else{if(!l){if(t.stateNode===null)throw Error(u(166));return cn(t),t.subtreeFlags&=-33554433,null}if(i=Ge.current,ur(t))om(t);else{var _=Xi(lt.current);switch(i){case 1:i=_.createElementNS("http://www.w3.org/2000/svg",a);break;case 2:i=_.createElementNS("http://www.w3.org/1998/Math/MathML",a);break;default:switch(a){case"svg":i=_.createElementNS("http://www.w3.org/2000/svg",a);break;case"math":i=_.createElementNS("http://www.w3.org/1998/Math/MathML",a);break;case"script":i=_.createElement("div"),i.innerHTML="<script><\/script>",i=i.removeChild(i.firstChild);break;case"select":i=typeof l.is=="string"?_.createElement("select",{is:l.is}):_.createElement("select"),l.multiple?i.multiple=!0:l.size&&(i.size=l.size);break;default:i=typeof l.is=="string"?_.createElement(a,{is:l.is}):_.createElement(a)}}i[En]=t,i[wn]=l;e:for(_=t.child;_!==null;){if(_.tag===5||_.tag===6)i.appendChild(_.stateNode);else if(_.tag!==4&&_.tag!==27&&_.child!==null){_.child.return=_,_=_.child;continue}if(_===t)break e;for(;_.sibling===null;){if(_.return===null||_.return===t)break e;_=_.return}_.sibling.return=_.return,_=_.sibling}t.stateNode=i;e:switch(nl(i,a,l),a){case"button":case"input":case"select":case"textarea":l=!!l.autoFocus;break e;case"img":l=!0;break e;default:l=!1}l&&os(t)}}return cn(t),t.subtreeFlags&=-33554433,tf(t,t.type,e===null?null:e.memoizedProps,t.pendingProps,n),null;case 6:if(e&&t.stateNode!=null)e.memoizedProps!==l&&os(t);else{if(typeof l!="string"&&t.stateNode===null)throw Error(u(166));if(e=lt.current,ur(t)){if(e=t.stateNode,n=t.memoizedProps,l=null,a=Qn,a!==null)switch(a.tag){case 27:case 5:l=a.memoizedProps}e[En]=t,e=!!(e.nodeValue===n||l!==null&&l.suppressHydrationWarning===!0||kg(e.nodeValue,n)),e||Ns(t,!0)}else e=Xi(e).createTextNode(l),e[En]=t,t.stateNode=e}return cn(t),null;case 31:if(n=t.memoizedState,e===null||e.memoizedState!==null){if(l=ur(t),n!==null){if(e===null){if(!l)throw Error(u(318));if(e=t.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(u(557));e[En]=t}else xa(),(t.flags&128)===0&&(t.memoizedState=null),t.flags|=4;cn(t),e=!1}else n=fd(),e!==null&&e.memoizedState!==null&&(e.memoizedState.hydrationErrors=n),e=!0;if(!e)return t.flags&256?($l(t),t):($l(t),null);if((t.flags&128)!==0)throw Error(u(558))}return cn(t),null;case 13:if(l=t.memoizedState,e===null||e.memoizedState!==null&&e.memoizedState.dehydrated!==null){if(a=ur(t),l!==null&&l.dehydrated!==null){if(e===null){if(!a)throw Error(u(318));if(a=t.memoizedState,a=a!==null?a.dehydrated:null,!a)throw Error(u(317));a[En]=t}else xa(),(t.flags&128)===0&&(t.memoizedState=null),t.flags|=4;cn(t),a=!1}else a=fd(),e!==null&&e.memoizedState!==null&&(e.memoizedState.hydrationErrors=a),a=!0;if(!a)return t.flags&256?($l(t),t):($l(t),null)}return $l(t),(t.flags&128)!==0?(t.lanes=n,t):(n=l!==null,e=e!==null&&e.memoizedState!==null,n&&(l=t.child,a=null,l.alternate!==null&&l.alternate.memoizedState!==null&&l.alternate.memoizedState.cachePool!==null&&(a=l.alternate.memoizedState.cachePool.pool),i=null,l.memoizedState!==null&&l.memoizedState.cachePool!==null&&(i=l.memoizedState.cachePool.pool),i!==a&&(l.flags|=2048)),n!==e&&n&&(t.child.flags|=8192),nu(t,t.updateQueue),cn(t),null);case 4:return Fe(),e===null&&Bf(t.stateNode.containerInfo),t.flags|=67108864,cn(t),null;case 10:return es(t.type),cn(t),null;case 19:if(Cd(t),l=t.memoizedState,l===null)return cn(t),null;if(a=(t.flags&128)!==0,i=l.rendering,i===null)if(a)Ai(l,!1);else{if(kn!==0||e!==null&&(e.flags&128)!==0)for(e=t.child;e!==null;){if(i=Yc(e),i!==null){for(t.flags|=128,Ai(l,!1),e=i.updateQueue,t.updateQueue=e,nu(t,e),t.subtreeFlags=0,e=n,n=t.child;n!==null;)Wt(n,e),n=n.sibling;return Mi(t,el.current&1|2),jt&&Jo(t,l.treeForkCount),t.child}e=e.sibling}l.tail!==null&&Ze()>mu&&(t.flags|=128,a=!0,Ai(l,!1),t.lanes=4194304)}else{if(!a)if(e=Yc(i),e!==null){if(t.flags|=128,a=!0,e=e.updateQueue,t.updateQueue=e,nu(t,e),Ai(l,!0),l.tail===null&&l.tailMode!=="collapsed"&&l.tailMode!=="visible"&&!i.alternate&&!jt)return cn(t),null}else 2*Ze()-l.renderingStartTime>mu&&n!==536870912&&(t.flags|=128,a=!0,Ai(l,!1),t.lanes=4194304);l.isBackwards?(i.sibling=t.child,t.child=i):(e=l.last,e!==null?e.sibling=i:t.child=i,l.last=i)}if(l.tail!==null){e=l.tail;e:{for(n=e;n!==null;){if(n.alternate!==null){n=!1;break e}n=n.sibling}n=!0}return l.rendering=e,l.tail=e.sibling,l.renderingStartTime=Ze(),e.sibling=null,i=el.current,i=a?i&1|2:i&1,l.tailMode==="visible"||l.tailMode==="collapsed"||!n||jt?Mi(t,i):(n=i,ye(Pn,t),ye(el,n),cl===null&&(cl=t)),jt&&Jo(t,l.treeForkCount),e}return cn(t),null;case 22:case 23:return $l(t),jd(),l=t.memoizedState!==null,e!==null?e.memoizedState!==null!==l&&(t.flags|=8192):l&&(t.flags|=8192),l?(n&536870912)!==0&&(t.flags&128)===0&&(cn(t),t.subtreeFlags&6&&(t.flags|=8192)):cn(t),n=t.updateQueue,n!==null&&nu(t,n.retryQueue),n=null,e!==null&&e.memoizedState!==null&&e.memoizedState.cachePool!==null&&(n=e.memoizedState.cachePool.pool),l=null,t.memoizedState!==null&&t.memoizedState.cachePool!==null&&(l=t.memoizedState.cachePool.pool),l!==n&&(t.flags|=2048),e!==null&&re(ja),null;case 24:return n=null,e!==null&&(n=e.memoizedState.cache),t.memoizedState.cache!==n&&(t.flags|=2048),es(Dn),cn(t),null;case 25:return null;case 30:return t.flags|=33554432,cn(t),null}throw Error(u(156,t.tag))}function Oy(e,t){switch(ud(t),t.tag){case 1:return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 3:return es(Dn),Fe(),e=t.flags,(e&65536)!==0&&(e&128)===0?(t.flags=e&-65537|128,t):null;case 26:case 27:case 5:return an(t),null;case 31:if(t.memoizedState!==null){if($l(t),t.alternate===null)throw Error(u(340));xa()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 13:if($l(t),e=t.memoizedState,e!==null&&e.dehydrated!==null){if(t.alternate===null)throw Error(u(340));xa()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 19:return Cd(t),e=t.flags,e&65536?(t.flags=e&-65537|128,e=t.memoizedState,e!==null&&(e.rendering=null,e.tail=null),t.flags|=4,t):null;case 4:return Fe(),null;case 10:return es(t.type),null;case 22:case 23:return $l(t),jd(),e!==null&&re(ja),e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 24:return es(Dn),null;case 25:return null;default:return null}}function Nh(e,t){switch(ud(t),t.tag){case 3:es(Dn),Fe();break;case 26:case 27:case 5:an(t);break;case 4:Fe();break;case 31:t.memoizedState!==null&&$l(t);break;case 13:$l(t);break;case 19:Cd(t);break;case 10:es(t.type);break;case 22:case 23:$l(t),jd(),e!==null&&re(ja);break;case 24:es(Dn)}}function Oi(e,t){try{var n=t.updateQueue,l=n!==null?n.lastEffect:null;if(l!==null){var a=l.next;n=a;do{if((n.tag&e)===e){l=void 0;var i=n.create,_=n.inst;l=i(),_.destroy=l}n=n.next}while(n!==a)}}catch(y){Zt(t,t.return,y)}}function Ls(e,t,n){try{var l=t.updateQueue,a=l!==null?l.lastEffect:null;if(a!==null){var i=a.next;l=i;do{if((l.tag&e)===e){var _=l.inst,y=_.destroy;if(y!==void 0){_.destroy=void 0,a=t;var M=n,L=y;try{L()}catch(X){Zt(a,M,X)}}}l=l.next}while(l!==i)}}catch(X){Zt(t,t.return,X)}}function Mh(e){var t=e.updateQueue;if(t!==null){var n=e.stateNode;try{pm(t,n)}catch(l){Zt(e,e.return,l)}}}function Eh(e,t,n){n.props=Ea(e.type,e.memoizedProps),n.state=e.memoizedState;try{n.componentWillUnmount()}catch(l){Zt(e,t,l)}}function To(e,t){try{var n=e.ref;if(n!==null){switch(e.tag){case 26:case 27:case 5:var l=e.stateNode;break;case 30:var a=e.stateNode,i=H(e.memoizedProps,a);(a.ref===null||a.ref.name!==i)&&(a.ref=Lg(i)),l=a.ref;break;case 7:if(e.stateNode===null){var _=new Il(e);b(e.child,!1,jx,_,void 0,void 0),e.stateNode=_}l=e.stateNode;break;default:l=e.stateNode}typeof n=="function"?e.refCleanup=n(l):n.current=l}}catch(y){Zt(e,t,y)}}function tl(e,t){var n=e.ref,l=e.refCleanup;if(n!==null)if(typeof l=="function")try{l()}catch(a){Zt(e,t,a)}finally{e.refCleanup=null,e=e.alternate,e!=null&&(e.refCleanup=null)}else if(typeof n=="function")try{n(null)}catch(a){Zt(e,t,a)}else n.current=null}function lu(e,t){if((e.tag===5||e.tag===27||e.tag===6)&&e.alternate===null&&t!==null)for(var n=0;n<t.length;n++)qg(e.stateNode,t[n])}function Th(e){for(var t=e.return;t!==null&&(lf(t)&&qg(e.stateNode,t.stateNode),!nf(t));)t=t.return}function Di(e){for(var t=e.return;t!==null&&(lf(t)&&kx(e.stateNode,t.stateNode),!nf(t));)t=t.return}function nf(e){return e.tag===5||e.tag===3||e.tag===27}function lf(e){return e&&e.tag===7&&e.stateNode!==null}function of(e){var t=e.type,n=e.memoizedProps,l=e.stateNode;try{e:switch(t){case"button":case"input":case"select":case"textarea":n.autoFocus&&l.focus();break e;case"img":n.src?l.src=n.src:n.srcSet&&(l.srcset=n.srcSet)}}catch(a){Zt(e,e.return,a)}}function sf(e,t,n){try{var l=e.stateNode;ax(l,e.type,n,t),l[wn]=t}catch(a){Zt(e,e.return,a)}}function Rh(e){return e.tag===5||e.tag===3||e.tag===26||e.tag===27&&Is(e.type)||e.tag===4}function af(e){e:for(;;){for(;e.sibling===null;){if(e.return===null||Rh(e.return))return null;e=e.return}for(e.sibling.return=e.return,e=e.sibling;e.tag!==5&&e.tag!==6&&e.tag!==18;){if(e.tag===27&&Is(e.type)||e.flags&2||e.child===null||e.tag===4)continue e;e.child.return=e,e=e.child}if(!(e.flags&2))return e.stateNode}}function rf(e,t,n,l){var a=e.tag;if(a===5||a===6)a=e.stateNode,t?(n.nodeType===9?n.body:n.nodeName==="HTML"?n.ownerDocument.body:n).insertBefore(a,t):(t=n.nodeType===9?n.body:n.nodeName==="HTML"?n.ownerDocument.body:n,t.appendChild(a),n=n._reactRootContainer,n!=null||t.onclick!==null||(t.onclick=no)),lu(e,l),yt=!0;else if(a!==4&&(a===27&&(lu(e,l),l=null,Is(e.type)&&(n=e.stateNode,t=null)),e=e.child,e!==null))for(rf(e,t,n,l),e=e.sibling;e!==null;)rf(e,t,n,l),e=e.sibling}function ou(e,t,n,l){var a=e.tag;if(a===5||a===6)a=e.stateNode,t?n.insertBefore(a,t):n.appendChild(a),lu(e,l),yt=!0;else if(a!==4&&(a===27&&(lu(e,l),l=null,Is(e.type)&&(n=e.stateNode)),e=e.child,e!==null))for(ou(e,t,n,l),e=e.sibling;e!==null;)ou(e,t,n,l),e=e.sibling}function zh(e){var t=e.stateNode,n=e.memoizedProps;try{for(var l=e.type,a=t.attributes;a.length;)t.removeAttributeNode(a[0]);nl(t,l,n),t[En]=e,t[wn]=n}catch(i){Zt(e,e.return,i)}}var su=!1,Ul=null;function Ah(e){(e.tag===30||(e.subtreeFlags&33554432)!==0)&&(su=!0)}var Ro=null;function Oh(){var e=Ro;return Ro=null,e}var Nl=0;function yr(e,t,n,l,a){return Nl=0,Dh(e.child,t,n,l,a)}function Dh(e,t,n,l,a){for(var i=!1;e!==null;){if(e.tag===5){var _=e.stateNode;if(l!==null){var y=Qf(_);l.push(y),y.view&&(i=!0)}else i||Qf(_).view&&(i=!0);su=!0,Og(_,Nl===0?t:t+"_"+Nl,n),Nl++}else(e.tag!==22||e.memoizedState===null)&&(e.tag===30&&a||Dh(e.child,t,n,l,a)&&(i=!0));e=e.sibling}return i}function zo(e,t){for(;e!==null;)e.tag===5?Dg(e.stateNode,e.memoizedProps):(e.tag!==22||e.memoizedState===null)&&(e.tag===30&&t||zo(e.child,t)),e=e.sibling}function au(e){if((e.subtreeFlags&18874368)!==0)for(e=e.child;e!==null;){if((e.tag!==22||e.memoizedState===null)&&(au(e),e.tag===30&&(e.flags&18874368)!==0&&e.stateNode.paired)){var t=e.memoizedProps;if(t.name==null||t.name==="auto")throw Error(u(544));var n=t.name;t=ce(t.default,t.share),t!=="none"&&(yr(e,n,t,null,!1)||zo(e.child,!1))}e=e.sibling}}function cf(e,t){if(e.tag===30){var n=e.stateNode,l=e.memoizedProps,a=H(l,n),i=ce(l.default,n.paired?l.share:l.enter);i!=="none"?yr(e,a,i,null,!1)?(au(e),n.paired||t||Cr(e,l.onEnter)):zo(e.child,!1):au(e)}else if((e.subtreeFlags&33554432)!==0)for(e=e.child;e!==null;)cf(e,t),e=e.sibling;else au(e)}function uf(e){if(Ul!==null&&Ul.size!==0){var t=Ul;if((e.subtreeFlags&18874368)!==0)for(e=e.child;e!==null;){if(e.tag!==22||e.memoizedState===null){if(e.tag===30&&(e.flags&18874368)!==0){var n=e.memoizedProps,l=n.name;if(l!=null&&l!=="auto"){var a=t.get(l);if(a!==void 0){var i=ce(n.default,n.share);if(i!=="none"&&(yr(e,l,i,null,!1)?(i=e.stateNode,a.paired=i,i.paired=a,Cr(e,n.onShare)):zo(e.child,!1)),t.delete(l),t.size===0)break}}}uf(e)}e=e.sibling}}}function df(e){if(e.tag===30){var t=e.memoizedProps,n=H(t,e.stateNode),l=Ul!==null?Ul.get(n):void 0,a=ce(t.default,l!==void 0?t.share:t.exit);a!=="none"&&(yr(e,n,a,null,!1)?l!==void 0?(a=e.stateNode,l.paired=a,a.paired=l,Ul.delete(n),Cr(e,t.onShare)):Cr(e,t.onExit):zo(e.child,!1)),Ul!==null&&uf(e)}else if((e.subtreeFlags&33554432)!==0)for(e=e.child;e!==null;)df(e),e=e.sibling;else Ul!==null&&uf(e)}function Lh(e){for(e=e.child;e!==null;){if(e.tag===30){var t=e.memoizedProps,n=H(t,e.stateNode);t=ce(t.default,t.update),e.flags&=-5,t!=="none"&&yr(e,n,t,e.memoizedState=[],!1)}else(e.subtreeFlags&33554432)!==0&&Lh(e);e=e.sibling}}function ff(e){if((e.subtreeFlags&18874368)!==0)for(e=e.child;e!==null;){if(e.tag!==22||e.memoizedState===null){if(e.tag===30&&(e.flags&18874368)!==0){var t=e.stateNode;t.paired!==null&&(t.paired=null,zo(e.child,!1))}ff(e)}e=e.sibling}}function ru(e){if(e.tag===30)e.stateNode.paired=null,zo(e.child,!1),ff(e);else if((e.subtreeFlags&33554432)!==0)for(e=e.child;e!==null;)ru(e),e=e.sibling;else ff(e)}function Bh(e){for(e=e.child;e!==null;)e.tag===30?zo(e.child,!1):(e.subtreeFlags&33554432)!==0&&Bh(e),e=e.sibling}function _f(e,t,n,l,a,i,_){for(var y=!1;t!==null;){if(t.tag===5){var M=t.stateNode;if(i!==null&&Nl<i.length){var L=i[Nl],X=Qf(M);(L.view||X.view)&&(y=!0);var J;if(J=(e.flags&4)===0)if(X.clip)J=!0;else{J=L.rect;var A=X.rect;J=J.y!==A.y||J.x!==A.x||J.height!==A.height||J.width!==A.width}J&&(e.flags|=4),X.abs?X=!L.abs:(L=L.rect,X=X.rect,X=L.height!==X.height||L.width!==X.width),X&&(e.flags|=32)}else e.flags|=32;(e.flags&4)!==0&&Og(M,Nl===0?n:n+"_"+Nl,a),y&&(e.flags&4)!==0||(Ro===null&&(Ro=[]),Ro.push(M,Nl===0?l:l+"_"+Nl,t.memoizedProps)),Nl++}else(t.tag!==22||t.memoizedState===null)&&(t.tag===30&&_?e.flags|=t.flags&32:_f(e,t.child,n,l,a,i,_)&&(y=!0));t=t.sibling}return y}function $h(e,t){for(e=e.child;e!==null;){if(e.tag===30){var n=e.memoizedProps,l=e.stateNode,a=H(n,l),i=ce(n.default,n.update),_;_=e.memoizedState,e.memoizedState=null,l=e;var y=e.child;Nl=0,a=_f(l,y,a,a,i,_,!1),(e.flags&4)!==0&&a&&Cr(e,n.onUpdate)}else(e.subtreeFlags&33554432)!==0&&$h(e);e=e.sibling}}var Gn=!1,Vt=!1,Ao=!1,mf=!1,Uh=typeof WeakSet=="function"?WeakSet:Set,Vn=null,Oo=!1,Li=!1,iu=!1,hf=!1;function Dy(e,t,n){if(e=e.containerInfo,Hf=Lr,e=hi(e),rr(e)){if("selectionStart"in e)var l={start:e.selectionStart,end:e.selectionEnd};else e:{l=(l=e.ownerDocument)&&l.defaultView||window;var a=l.getSelection&&l.getSelection();if(a&&a.rangeCount!==0){l=a.anchorNode;var i=a.anchorOffset,_=a.focusNode;a=a.focusOffset;try{l.nodeType,_.nodeType}catch{l=null;break e}var y=0,M=-1,L=-1,X=0,J=0,A=e,q=null;t:for(;;){for(var ve;A!==l||i!==0&&A.nodeType!==3||(M=y+i),A!==_||a!==0&&A.nodeType!==3||(L=y+a),A.nodeType===3&&(y+=A.nodeValue.length),(ve=A.firstChild)!==null;)q=A,A=ve;for(;;){if(A===e)break t;if(q===l&&++X===i&&(M=y),q===_&&++J===a&&(L=y),(ve=A.nextSibling)!==null)break;A=q,q=A.parentNode}A=ve}l=M===-1||L===-1?null:{start:M,end:L}}else l=null}l=l||{start:0,end:0}}else l=null;for(Yf={focusedElem:e,selectionRange:l},Lr=!1,n=(n&335544064)===n,Vn=t,t=n?9270:1024;Vn!==null;){if(e=Vn,n&&(l=e.deletions,l!==null))for(i=0;i<l.length;i++)n&&df(l[i]);if(e.alternate===null&&(e.flags&2)!==0)n&&Ah(e),cu(n);else{if(e.tag===22){if(l=e.alternate,e.memoizedState!==null){l!==null&&l.memoizedState===null&&n&&df(l),cu(n);continue}else if(l!==null&&l.memoizedState!==null){n&&Ah(e),cu(n);continue}}l=e.child,(e.subtreeFlags&t)!==0&&l!==null?(l.return=e,Vn=l):(n&&Lh(e),cu(n))}}Ul=null}function cu(e){for(;Vn!==null;){var t=Vn,n=e,l=t.alternate,a=t.flags;switch(t.tag){case 0:case 11:case 15:break;case 1:if((a&1024)!==0&&l!==null){n=void 0,a=l.memoizedProps,l=l.memoizedState;var i=t.stateNode;try{var _=Ea(t.type,a);n=i.getSnapshotBeforeUpdate(_,l),i.__reactInternalSnapshotBeforeUpdate=n}catch(y){Zt(t,t.return,y)}}break;case 3:if((a&1024)!==0){if(l=t.stateNode.containerInfo,n=l.nodeType,n===9)Ff(l);else if(n===1)switch(l.nodeName){case"HEAD":case"HTML":case"BODY":Ff(l);break;default:l.textContent=""}}break;case 5:case 26:case 27:case 6:case 4:case 17:break;case 30:n&&l!==null&&(n=H(l.memoizedProps,l.stateNode),a=t.memoizedProps,a=ce(a.default,a.update),a!=="none"&&yr(l,n,a,l.memoizedState=[],!0));break;default:if((a&1024)!==0)throw Error(u(163))}if(l=t.sibling,l!==null){l.return=t.return,Vn=l;break}Vn=t.return}}function Hh(e,t,n){var l=n.flags;switch(n.tag){case 0:case 11:case 15:Do(e,n),l&4&&Oi(5,n);break;case 1:if(Do(e,n),l&4)if(e=n.stateNode,t===null)try{e.componentDidMount()}catch(_){Zt(n,n.return,_)}else{var a=Ea(n.type,t.memoizedProps);t=t.memoizedState;try{e.componentDidUpdate(a,t,e.__reactInternalSnapshotBeforeUpdate)}catch(_){Zt(n,n.return,_)}}l&64&&Mh(n),l&512&&To(n,n.return);break;case 3:if(Do(e,n),l&64&&(e=n.updateQueue,e!==null)){if(t=null,n.child!==null)switch(n.child.tag){case 27:case 5:t=n.child.stateNode;break;case 1:t=n.child.stateNode}try{pm(e,t)}catch(_){Zt(n,n.return,_)}}break;case 27:t===null&&l&4&&zh(n);case 26:case 5:Do(e,n),t===null&&l&4&&of(n),l&512&&To(n,n.return);break;case 12:Do(e,n);break;case 31:Do(e,n),l&4&&Xh(e,n);break;case 13:Do(e,n),l&4&&Qh(e,n),l&64&&(e=n.memoizedState,e!==null&&(e=e.dehydrated,e!==null&&(n=Vy.bind(null,n),Mx(e,n))));break;case 22:if(l=n.memoizedState!==null||Gn,!l){var i=t!==null&&t.memoizedState!==null||Vt;t=Gn,a=Vt,Gn=l,(Vt=i)&&!a?(l=2,(n.subtreeFlags&8772)!==0&&(l|=1),uo(e,n,l)):Do(e,n),Gn=t,Vt=a}break;case 30:Do(e,n),l&512&&To(n,n.return);break;case 7:l&512&&To(n,n.return);default:Do(e,n)}}function gf(e,t){for(e=e.child;e!==null;)Yh(e,t),e=e.sibling}function Yh(e,t){switch(e.tag){case 5:case 26:try{var n=e.stateNode;if(t){var l=n.style;typeof l.setProperty=="function"?l.setProperty("display","none","important"):l.display="none"}else{var a=e.stateNode,i=e.memoizedProps.style,_=i!=null&&i.hasOwnProperty("display")?i.display:null;a.style.display=_==null||typeof _=="boolean"?"":(""+_).trim()}}catch(M){Zt(e,e.return,M)}pf(e,t);break;case 6:try{e.stateNode.nodeValue=t?"":e.memoizedProps,yt=!0}catch(M){Zt(e,e.return,M)}break;case 18:try{var y=e.stateNode;t?Ag(y,!0):Ag(e.stateNode,!1)}catch(M){Zt(e,e.return,M)}break;case 22:case 23:e.memoizedState===null&&gf(e,t);break;default:gf(e,t)}}function pf(e,t){if(e.subtreeFlags&67108864)for(e=e.child;e!==null;){e:{var n=e,l=t;switch(n.tag){case 4:Yh(n,l);break e;case 22:n.memoizedState===null&&pf(n,l);break e;default:pf(n,l)}}e=e.sibling}}function qh(e){var t=e.alternate;t!==null&&(e.alternate=null,qh(t)),e.child=null,e.deletions=null,e.sibling=null,e.tag===5&&(t=e.stateNode,t!==null&&ea(t)),e.stateNode=null,e.return=null,e.dependencies=null,e.memoizedProps=null,e.memoizedState=null,e.pendingProps=null,e.stateNode=null,e.updateQueue=null}var hn=null,Ml=!1;function io(e,t,n){for(n=n.child;n!==null;)Ih(e,t,n),n=n.sibling}function Ih(e,t,n){if(Je&&typeof Je.onCommitFiberUnmount=="function")try{Je.onCommitFiberUnmount(vt,n)}catch{}switch(n.tag){case 26:Vt||tl(n,t),io(e,t,n),n.memoizedState?n.memoizedState.count--:n.stateNode&&!Vt&&(n=n.stateNode,n.parentNode.removeChild(n));break;case 27:Vt||tl(n,t),Di(n);var l=hn,a=Ml;Is(n.type)&&(hn=n.stateNode,Ml=!1),io(e,t,n),Vg(n.stateNode,n.type,n.memoizedProps),hn=l,Ml=a;break;case 5:Vt||tl(n,t),Di(n);case 6:if(n.tag===6&&Di(n),l=hn,a=Ml,hn=null,io(e,t,n),hn=l,Ml=a,hn!==null)if(Ml)try{(hn.nodeType===9?hn.body:hn.nodeName==="HTML"?hn.ownerDocument.body:hn).removeChild(n.stateNode),yt=!0}catch(i){Zt(n,t,i)}else try{hn.removeChild(n.stateNode),yt=!0}catch(i){Zt(n,t,i)}break;case 18:hn!==null&&(Ml?(e=hn,zg(e.nodeType===9?e.body:e.nodeName==="HTML"?e.ownerDocument.body:e,n.stateNode),Br(e)):zg(hn,n.stateNode));break;case 4:l=hn,a=Ml,hn=n.stateNode.containerInfo,Ml=!0,io(e,t,n),hn=l,Ml=a;break;case 0:case 11:case 14:case 15:Ls(2,n,t),Vt||Ls(4,n,t),io(e,t,n);break;case 1:Vt||(tl(n,t),l=n.stateNode,typeof l.componentWillUnmount=="function"&&Eh(n,t,l)),io(e,t,n);break;case 21:io(e,t,n);break;case 22:Vt=(l=Vt)||n.memoizedState!==null,io(e,t,n),Vt=l;break;case 30:tl(n,t),io(e,t,n);break;case 7:Vt||tl(n,t),io(e,t,n);break;default:io(e,t,n)}}function Xh(e,t){if(t.memoizedState===null&&(e=t.alternate,e!==null&&(e=e.memoizedState,e!==null))){e=e.dehydrated;try{Br(e)}catch(n){Zt(t,t.return,n)}}}function Qh(e,t){if(t.memoizedState===null&&(e=t.alternate,e!==null&&(e=e.memoizedState,e!==null&&(e=e.dehydrated,e!==null))))try{Br(e)}catch(n){Zt(t,t.return,n)}}function Ly(e){switch(e.tag){case 31:case 13:case 19:var t=e.stateNode;return t===null&&(t=e.stateNode=new Uh),t;case 22:return e=e.stateNode,t=e._retryCache,t===null&&(t=e._retryCache=new Uh),t;default:throw Error(u(435,e.tag))}}function uu(e,t){var n=Ly(e);t.forEach(function(l){if(!n.has(l)){n.add(l);var a=Fy.bind(null,e,l);l.then(a,a)}})}function yl(e,t,n){var l=t.deletions;if(l!==null)for(var a=0;a<l.length;a++){var i=l[a],_=e,y=t,M=y;e:for(;M!==null;){switch(M.tag){case 27:if(Is(M.type)){hn=M.stateNode,Ml=!1;break e}break;case 5:hn=M.stateNode,Ml=!1;break e;case 3:case 4:hn=M.stateNode.containerInfo,Ml=!0;break e}M=M.return}if(hn===null)throw Error(u(160));Ih(_,y,i),hn=null,Ml=!1,_=i.alternate,_!==null&&(_.return=null),i.return=null}if(t.subtreeFlags&13886)for(t=t.child;t!==null;)Gh(t,e,n),t=t.sibling}var co=null;function Gh(e,t,n){var l=e.alternate,a=e.flags;switch(e.tag){case 0:case 11:case 14:case 15:if(a&4&&(l=e.updateQueue,l=l!==null?l.events:null,l!==null))for(var i=0;i<l.length;i++){var _=l[i];_.ref.impl=_.nextImpl}yl(t,e,n),xl(e),a&4&&(Ls(3,e,e.return),Oi(3,e),Ls(5,e,e.return));break;case 1:yl(t,e,n),xl(e),a&512&&(Vt||l===null||tl(l,l.return)),a&64&&Gn&&(e=e.updateQueue,e!==null&&(t=e.callbacks,t!==null&&(n=e.shared.hiddenCallbacks,e.shared.hiddenCallbacks=n===null?t:n.concat(t))));break;case 26:if(i=co,yl(t,e,n),xl(e),a&512&&(Vt||l===null||tl(l,l.return)),a&4)if(a=l!==null?l.memoizedState:null,n=e.memoizedState,l===null)if(n===null)if(e.stateNode===null)if(Gn)e.stateNode=Eg(e.type,e.memoizedProps,t.containerInfo,e);else{e:{t=e.type,n=e.memoizedProps,a=i.ownerDocument||i;t:switch(t){case"title":l=a.getElementsByTagName("title")[0],(!l||l[Uo]||l[En]||l.namespaceURI==="http://www.w3.org/2000/svg"||l.hasAttribute("itemprop"))&&(l=a.createElement(t),a.head.insertBefore(l,a.querySelector("head > title"))),nl(l,t,n),l[En]=e,Tn(l),t=l;break e;case"link":if(i=Pg("link","href",a).get(t+(n.href||""))){for(_=0;_<i.length;_++)if(l=i[_],l.getAttribute("href")===(n.href==null||n.href===""?null:n.href)&&l.getAttribute("rel")===(n.rel==null?null:n.rel)&&l.getAttribute("title")===(n.title==null?null:n.title)&&l.getAttribute("crossorigin")===(n.crossOrigin==null?null:n.crossOrigin)){i.splice(_,1);break t}}l=a.createElement(t),nl(l,t,n),a.head.appendChild(l);break;case"meta":if(i=Pg("meta","content",a).get(t+(n.content||""))){for(_=0;_<i.length;_++)if(l=i[_],l.getAttribute("content")===(n.content==null?null:""+n.content)&&l.getAttribute("name")===(n.name==null?null:n.name)&&l.getAttribute("property")===(n.property==null?null:n.property)&&l.getAttribute("http-equiv")===(n.httpEquiv==null?null:n.httpEquiv)&&l.getAttribute("charset")===(n.charSet==null?null:n.charSet)){i.splice(_,1);break t}}l=a.createElement(t),nl(l,t,n),a.head.appendChild(l);break;default:throw Error(u(468,t))}l[En]=e,Tn(l),t=l}e.stateNode=t}else Gn||t_(i,e.type,e.stateNode);else e.stateNode=Jg(i,n,e.memoizedProps);else a!==n?(a===null?(t=l.stateNode,t===null||Vt||t.parentNode.removeChild(t)):a.count--,n===null?Gn||t_(i,e.type,e.stateNode):Jg(i,n,e.memoizedProps)):n===null&&e.stateNode!==null&&sf(e,e.memoizedProps,l.memoizedProps);break;case 27:yl(t,e,n),xl(e),a&512&&(Vt||l===null||tl(l,l.return)),l!==null&&a&4&&sf(e,e.memoizedProps,l.memoizedProps);break;case 5:if(i=Ao,Ao=!1,yl(t,e,n),Ao=i,xl(e),a&512&&(Vt||l===null||tl(l,l.return)),e.flags&32){t=e.stateNode;try{vo(t,""),yt=!0}catch(X){Zt(e,e.return,X)}}a&4&&e.stateNode!=null&&(t=e.memoizedProps,sf(e,t,l!==null?l.memoizedProps:t)),a&1024&&(mf=!0);break;case 6:if(yl(t,e,n),xl(e),a&4){if(e.stateNode===null)throw Error(u(162));t=e.memoizedProps,n=e.stateNode;try{n.nodeValue=t,yt=!0}catch(X){Zt(e,e.return,X)}}break;case 3:if(yt=!1,ku=null,i=co,co=Qi(t.containerInfo),yl(t,e,n),co=i,xl(e),a&4&&l!==null&&l.memoizedState.isDehydrated)try{Br(t.containerInfo)}catch(X){Zt(e,e.return,X)}mf&&(mf=!1,Vh(e)),yt=!1;break;case 4:a=Ao,Ao=Gn,l=gc(),i=co,co=Qi(e.stateNode.containerInfo),yl(t,e,n),xl(e),co=i,yt&&Li&&(iu=!0),yt=l,Ao=a;break;case 12:yl(t,e,n),xl(e);break;case 31:yl(t,e,n),xl(e),a&4&&(t=e.updateQueue,t!==null&&(e.updateQueue=null,uu(e,t)));break;case 13:yl(t,e,n),xl(e),e.child.flags&8192&&e.memoizedState!==null!=(l!==null&&l.memoizedState!==null)&&(_u=Ze()),a&4&&(t=e.updateQueue,t!==null&&(e.updateQueue=null,uu(e,t)));break;case 22:i=e.memoizedState!==null,_=l!==null&&l.memoizedState!==null;var y=Gn,M=Vt,L=Ao;Gn=y||i,Ao=L||i,Vt=M||_,yl(t,e,n),Vt=M,Ao=L,Gn=y,xl(e),a&8192&&(t=e.stateNode,t._visibility=i?t._visibility&-2:t._visibility|1,!i||l===null||_||Gn||Vt||(t=_||Vt,n=Gn,l=Vt,Gn=i||Gn,Vt=t,Bs(e,2),Gn=n,Vt=l),!i&&Ao||gf(e,i)),a&4&&(t=e.updateQueue,t!==null&&(n=t.retryQueue,n!==null&&(t.retryQueue=null,uu(e,n))));break;case 19:yl(t,e,n),xl(e),a&4&&(t=e.updateQueue,t!==null&&(e.updateQueue=null,uu(e,t)));break;case 30:a&512&&(Vt||l===null||tl(l,l.return)),a=gc(),i=Li,_=(n&335544064)===n,y=e.memoizedProps,Li=_&&ce(y.default,y.update)!=="none",yl(t,e,n),xl(e),_&&l!==null&&yt&&(e.flags|=4),Li=i,yt=a;break;case 21:break;case 7:a&512&&(Vt||l===null||tl(l,l.return)),l&&l.stateNode!==null&&(l.stateNode._fragmentFiber=e);default:yl(t,e,n),xl(e)}}function xl(e){var t=e.flags;if(t&2){try{for(var n,l=e.return;l!==null;){if(Rh(l)){n=l;break}l=l.return}l=null;for(var a=e.return;a!==null;){if(lf(a)){var i=a.stateNode;l===null?l=[i]:l.push(i)}if(nf(a))break;a=a.return}var _=l;if(n==null)throw Error(u(160));switch(n.tag){case 27:var y=n.stateNode,M=af(e);ou(e,M,y,_);break;case 5:var L=n.stateNode;n.flags&32&&(vo(L,""),n.flags&=-33);var X=af(e);ou(e,X,L,_);break;case 3:case 4:var J=n.stateNode.containerInfo,A=af(e);rf(e,A,J,_);break;default:throw Error(u(161))}}catch(q){Zt(e,e.return,q)}e.flags&=-3}t&4096&&(e.flags&=-4097)}function Vh(e){if(e.subtreeFlags&1024)for(e=e.child;e!==null;){var t=e;Vh(t),t.tag===5&&t.flags&1024&&(t=t.stateNode,Lr=!0,t.reset(),Lr=!1),e=e.sibling}}function xr(e,t){if(t.subtreeFlags&9270)for(t=t.child;t!==null;)Fh(t,e),t=t.sibling;else $h(t)}function Fh(e,t){var n=e.alternate;if(n===null)cf(e,!1);else switch(e.tag){case 3:if(hf=Oo=!1,Oh(),xr(t,e),!Oo&&!iu){if(e=Ro,e!==null)for(var l=0;l<e.length;l+=3){n=e[l];var a=e[l+1];Dg(n,e[l+2]),n=n.ownerDocument.documentElement,n!==null&&n.animate({opacity:[0,0],pointerEvents:["none","none"]},{duration:0,fill:"forwards",pseudoElement:"::view-transition-group("+a+")"})}e=t.containerInfo,e=e.nodeType===9?e.documentElement:e.ownerDocument.documentElement,e!==null&&e.style.viewTransitionName===""&&(e.style.viewTransitionName="none",e.animate({opacity:[0,0],pointerEvents:["none","none"]},{duration:0,fill:"forwards",pseudoElement:"::view-transition-group(root)"}),e.animate({width:[0,0],height:[0,0]},{duration:0,fill:"forwards",pseudoElement:"::view-transition"})),hf=!0}Ro=null;break;case 5:xr(t,e);break;case 4:l=Oo,Oo=!1,xr(t,e),Oo&&(iu=!0),Oo=l;break;case 22:e.memoizedState===null&&(n.memoizedState!==null?cf(e,!1):xr(t,e));break;case 30:l=Oo,a=Oh(),Oo=!1,xr(t,e),Oo&&(e.flags|=4);var i=e.memoizedProps,_=e.stateNode;t=H(i,_),_=H(n.memoizedProps,_);var y=ce(i.default,i.update);y==="none"?t=!1:(i=n.memoizedState,n.memoizedState=null,n=e.child,Nl=0,t=_f(e,n,t,_,y,i,!0),Nl!==(i===null?0:i.length)&&(e.flags|=32)),(e.flags&4)!==0&&t?(Cr(e,e.memoizedProps.onUpdate),Ro=a):a!==null&&(a.push.apply(a,Ro),Ro=a),Oo=(e.flags&32)!==0?!0:l;break;default:xr(t,e)}}function Do(e,t){if(t.subtreeFlags&8772)for(t=t.child;t!==null;)Hh(e,t.alternate,t),t=t.sibling}function Bs(e,t){for(e=e.child;e!==null;){var n=e,l=t;switch(n.tag){case 0:case 11:case 14:case 15:Ls(4,n,n.return),Bs(n,l);break;case 1:tl(n,n.return);var a=n.stateNode;typeof a.componentWillUnmount=="function"&&Eh(n,n.return,a),Bs(n,l);break;case 27:(l&2)!==0&&Vg(n.stateNode,n.type,n.memoizedProps);case 5:tl(n,n.return),n.tag!==5&&n.tag!==27||Di(n),Bs(n,l);break;case 6:Di(n);break;case 26:tl(n,n.return),a=n.stateNode,n.memoizedState!==null||a===null||Vt||a.parentNode.removeChild(a),Bs(n,l);break;case 22:n.memoizedState===null&&Bs(n,l);break;case 30:tl(n,n.return),Bs(n,l);break;case 7:tl(n,n.return);default:Bs(n,l)}e=e.sibling}}function uo(e,t,n){for(n=(t.subtreeFlags&8772)!==0?n:n&-2,t=t.child;t!==null;){var l=t.alternate,a=e,i=t,_=i.flags,y=(n&1)!==0;switch(i.tag){case 0:case 11:case 15:uo(a,i,n),Oi(4,i);break;case 1:if(uo(a,i,n),l=i,a=l.stateNode,typeof a.componentDidMount=="function")try{a.componentDidMount()}catch(X){Zt(l,l.return,X)}if(l=i,a=l.updateQueue,a!==null){var M=l.stateNode;try{var L=a.shared.hiddenCallbacks;if(L!==null)for(a.shared.hiddenCallbacks=null,a=0;a<L.length;a++)gm(L[a],M)}catch(X){Zt(l,l.return,X)}}y&&_&64&&Mh(i),To(i,i.return);break;case 27:(n&2)!==0&&zh(i);case 5:i.tag!==5&&i.tag!==27||Th(i),uo(a,i,n),y&&l===null&&_&4&&of(i),To(i,i.return);break;case 6:Th(i);break;case 26:M=i.stateNode,i.memoizedState!==null||M===null||Gn||t_(Qi(M.ownerDocument),i.type,M),uo(a,i,n),y&&l===null&&_&4&&of(i),To(i,i.return);break;case 12:uo(a,i,n);break;case 31:uo(a,i,n),y&&_&4&&Xh(a,i);break;case 13:uo(a,i,n),y&&_&4&&Qh(a,i);break;case 22:i.memoizedState===null&&uo(a,i,n),To(i,i.return);break;case 30:uo(a,i,n),To(i,i.return);break;case 7:To(i,i.return);default:uo(a,i,n)}t=t.sibling}}function yf(e,t){var n=null;e!==null&&e.memoizedState!==null&&e.memoizedState.cachePool!==null&&(n=e.memoizedState.cachePool.pool),e=null,t.memoizedState!==null&&t.memoizedState.cachePool!==null&&(e=t.memoizedState.cachePool.pool),e!==n&&(e!=null&&e.refCount++,n!=null&&vi(n))}function xf(e,t){e=null,t.alternate!==null&&(e=t.alternate.memoizedState.cache),t=t.memoizedState.cache,t!==e&&(t.refCount++,e!=null&&vi(e))}function Zl(e,t,n,l){var a=(n&335544064)===n;if(t.subtreeFlags&(a?10262:10256))for(t=t.child;t!==null;)Wh(e,t,n,l),t=t.sibling;else a&&Bh(t)}function Wh(e,t,n,l){var a=(n&335544064)===n;a&&t.alternate===null&&t.return!==null&&t.return.alternate!==null&&ru(t);var i=t.flags;switch(t.tag){case 0:case 11:case 15:Zl(e,t,n,l),i&2048&&Oi(9,t);break;case 1:Zl(e,t,n,l);break;case 3:Zl(e,t,n,l),a&&hf&&(e=e.containerInfo,e=e.nodeType===9?e.body:e.nodeName==="HTML"?e.ownerDocument.body:e,e.style.viewTransitionName==="root"&&(e.style.viewTransitionName=""),e=e.ownerDocument.documentElement,e!==null&&e.style.viewTransitionName==="none"&&(e.style.viewTransitionName="")),i&2048&&(i=null,t.alternate!==null&&(i=t.alternate.memoizedState.cache),t=t.memoizedState.cache,t!==i&&(t.refCount++,i!=null&&vi(i)));break;case 12:if(i&2048){Zl(e,t,n,l),i=t.stateNode;try{var _=t.memoizedProps,y=_.id,M=_.onPostCommit;typeof M=="function"&&M(y,t.alternate===null?"mount":"update",i.passiveEffectDuration,-0)}catch(L){Zt(t,t.return,L)}}else Zl(e,t,n,l);break;case 31:Zl(e,t,n,l);break;case 13:Zl(e,t,n,l);break;case 23:break;case 22:_=t.stateNode,y=t.alternate,t.memoizedState!==null?(a&&y!==null&&y.memoizedState===null&&ru(y),_._visibility&2?Zl(e,t,n,l):Bi(e,t)):(a&&y!==null&&y.memoizedState!==null&&ru(t),_._visibility&2?Zl(e,t,n,l):(_._visibility|=2,br(e,t,n,l,(t.subtreeFlags&10256)!==0||!1))),i&2048&&yf(y,t);break;case 24:Zl(e,t,n,l),i&2048&&xf(t.alternate,t);break;case 30:a&&(i=t.alternate,i!==null&&(zo(i.child,!0),zo(t.child,!0))),Zl(e,t,n,l);break;default:Zl(e,t,n,l)}}function br(e,t,n,l,a){for(a=a&&((t.subtreeFlags&10256)!==0||!1),t=t.child;t!==null;){var i=e,_=t,y=n,M=l,L=_.flags;switch(_.tag){case 0:case 11:case 15:br(i,_,y,M,a),Oi(8,_);break;case 23:break;case 22:var X=_.stateNode;_.memoizedState!==null?X._visibility&2?br(i,_,y,M,a):Bi(i,_):(X._visibility|=2,br(i,_,y,M,a)),a&&L&2048&&yf(_.alternate,_);break;case 24:br(i,_,y,M,a),a&&L&2048&&xf(_.alternate,_);break;default:br(i,_,y,M,a)}t=t.sibling}}function Bi(e,t){if(t.subtreeFlags&10256)for(t=t.child;t!==null;){var n=e,l=t,a=l.flags;switch(l.tag){case 22:Bi(n,l),a&2048&&yf(l.alternate,l);break;case 24:Bi(n,l),a&2048&&xf(l.alternate,l);break;default:Bi(n,l)}t=t.sibling}}var Ta=8192;function Ra(e,t,n){if(e.subtreeFlags&Ta)for(e=e.child;e!==null;)Kh(e,t,n),e=e.sibling}function Kh(e,t,n){switch(e.tag){case 26:Ra(e,t,n),e.flags&Ta&&(e.memoizedState!==null?qx(n,co,e.memoizedState,e.memoizedProps):(e=e.stateNode,(t&335544128)===t&&l0(n,e)));break;case 5:Ra(e,t,n),e.flags&Ta&&(e=e.stateNode,(t&335544128)===t&&l0(n,e));break;case 3:case 4:var l=co;co=Qi(e.stateNode.containerInfo),Ra(e,t,n),co=l;break;case 22:e.memoizedState===null&&(l=e.alternate,l!==null&&l.memoizedState!==null?(l=Ta,Ta=16777216,Ra(e,t,n),Ta=l):Ra(e,t,n));break;case 30:if((e.flags&Ta)!==0&&(l=e.memoizedProps.name,l!=null&&l!=="auto")){var a=e.stateNode;a.paired=null,Ul===null&&(Ul=new Map),Ul.set(l,a)}Ra(e,t,n);break;default:Ra(e,t,n)}}function Zh(e){var t=e.alternate;if(t!==null&&(e=t.child,e!==null)){t.child=null;do t=e.sibling,e.sibling=null,e=t;while(e!==null)}}function $i(e){var t=e.deletions;if((e.flags&16)!==0){if(t!==null)for(var n=0;n<t.length;n++){var l=t[n];Vn=l,Ph(l,e)}Zh(e)}if(e.subtreeFlags&10256)for(e=e.child;e!==null;)Jh(e),e=e.sibling}function Jh(e){switch(e.tag){case 0:case 11:case 15:$i(e),e.flags&2048&&Ls(9,e,e.return);break;case 3:$i(e);break;case 12:$i(e);break;case 22:var t=e.stateNode;e.memoizedState!==null&&t._visibility&2&&(e.return===null||e.return.tag!==13)?(t._visibility&=-3,du(e)):$i(e);break;default:$i(e)}}function du(e){var t=e.deletions;if((e.flags&16)!==0){if(t!==null)for(var n=0;n<t.length;n++){var l=t[n];Vn=l,Ph(l,e)}Zh(e)}for(e=e.child;e!==null;){switch(t=e,t.tag){case 0:case 11:case 15:Ls(8,t,t.return),du(t);break;case 22:n=t.stateNode,n._visibility&2&&(n._visibility&=-3,du(t));break;default:du(t)}e=e.sibling}}function Ph(e,t){for(;Vn!==null;){var n=Vn;switch(n.tag){case 0:case 11:case 15:Ls(8,n,t);break;case 23:case 22:if(n.memoizedState!==null&&n.memoizedState.cachePool!==null){var l=n.memoizedState.cachePool.pool;l!=null&&l.refCount++}break;case 24:vi(n.memoizedState.cache)}if(l=n.child,l!==null)l.return=n,Vn=l;else e:for(n=e;Vn!==null;){l=Vn;var a=l.sibling,i=l.return;if(qh(l),l===n){Vn=null;break e}if(a!==null){a.return=i,Vn=a;break e}Vn=i}}}var By={getCacheForType:function(e){var t=Jn(Dn),n=t.data.get(e);return n===void 0&&(n=e(),t.data.set(e,n)),n},cacheSignal:function(){return Jn(Dn).controller.signal}},$y=typeof WeakMap=="function"?WeakMap:Map,qt=0,sn=null,Mt=null,Rt=0,Kt=0,Hl=null,$s=!1,vr=!1,bf=!1,ss=0,kn=0,Us=0,za=0,fu=0,Yl=0,wr=0,Ui=null,El=null,vf=!1,_u=0,eg=0,mu=1/0,hu=null,Hs=null,xn=0,fo=null,Aa=null,Lo=0,wf=0,Sf=null,tg=null,Sr=null,jr=null,kr=null,Hi=0,gu=null;function ql(){return(qt&2)!==0&&Rt!==0?Rt&-Rt:Re.T!==null?Af():Kr()}function ng(){if(Yl===0)if((Rt&536870912)===0||jt){var e=Fn;Fn<<=1,(Fn&3932160)===0&&(Fn=262144),Yl=e}else Yl=536870912;return e=Pn.current,e!==null&&(e.flags|=32),Yl}function Cr(e,t){if(t!=null){var n=e.stateNode,l=n.ref;l===null&&(l=n.ref=Lg(H(e.memoizedProps,n))),jr===null&&(jr=[]),jr.push(t.bind(null,l))}}function Tl(e,t,n){(e===sn&&(Kt===2||Kt===9)||e.cancelPendingCommit!==null)&&(Nr(e,0),Ys(e,Rt,Yl,!1)),Ps(e,n),((qt&2)===0||e!==sn)&&(e===sn&&((qt&2)===0&&(za|=n),kn===4&&Ys(e,Rt,Yl,!1)),Bo(e))}function lg(e,t,n){if((qt&6)!==0)throw Error(u(327));var l=!n&&(t&127)===0&&(t&e.expiredLanes)===0||sl(e,t),a=l?Yy(e,t):kf(e,t,!0),i=l;do{if(a===0){vr&&!l&&Ys(e,t,0,!1);break}else{if(n=e.current.alternate,i&&!Uy(n)){a=kf(e,t,!1),i=!1;continue}if(a===2){if(i=t,e.errorRecoveryDisabledLanes&i)var _=0;else _=e.pendingLanes&-536870913,_=_!==0?_:_&536870912?536870912:0;if(_!==0){t=_;e:{var y=e;a=Ui;var M=y.current.memoizedState.isDehydrated;if(M&&(Nr(y,_).flags|=256),_=kf(y,_,!1),_!==2&&_!==6){if(bf&&!M){y.errorRecoveryDisabledLanes|=i,za|=i,a=4;break e}i=El,El=a,i!==null&&(El===null?El=i:El.push.apply(El,i))}a=_}if(i=!1,a!==2)continue}}if(a===1){Nr(e,0),Ys(e,t,0,!0);break}e:{switch(l=e,i=a,i){case 0:case 1:throw Error(u(345));case 4:if((t&4194048)!==t&&(t&62914560)!==t)break;case 6:Ys(l,t,Yl,!$s);break e;case 2:El=null;break;case 3:case 5:break;default:throw Error(u(329))}if((t&62914560)===t&&(a=_u+300-Ze(),10<a)){if(Ys(l,t,Yl,!$s),dl(l,0,!0)!==0)break e;Lo=t,l.timeoutHandle=Xf(og.bind(null,l,n,El,hu,vf,t,Yl,za,wr,$s,i,"Throttled",-0,0),a);break e}og(l,n,El,hu,vf,t,Yl,za,wr,$s,i,null,-0,0)}}break}while(!0);Bo(e)}function og(e,t,n,l,a,i,_,y,M,L,X,J,A,q){e.timeoutHandle=-1;var ve=t.subtreeFlags,Le=(i&335544064)===i;if(J=null,(Le||ve&8192||(ve&16785408)===16785408)&&(J={stylesheets:null,count:0,imgCount:0,imgBytes:0,suspenseyImages:[],waitingForImages:!0,waitingForViewTransition:!1,unsuspend:no},Ul=null,Kh(t,i,J),Le&&(ve=J,Le=e.containerInfo,Le=(Le.nodeType===9?Le:Le.ownerDocument).__reactViewTransition,Le!=null&&(ve.count++,ve.waitingForViewTransition=!0,ve=Fi.bind(ve),Le.finished.then(ve,ve))),ve=(i&62914560)===i?_u-Ze():(i&4194048)===i?eg-Ze():0,ve=Ix(J,ve),ve!==null)){Lo=i,e.cancelPendingCommit=ve(fg.bind(null,e,t,i,n,l,a,_,y,M,L,X,J,null,A,q)),Ys(e,i,_,!L);return}fg(e,t,i,n,l,a,_,y,M,L,X,J)}function Uy(e){for(var t=e;;){var n=t.tag;if((n===0||n===11||n===15)&&t.flags&16384&&(n=t.updateQueue,n!==null&&(n=n.stores,n!==null)))for(var l=0;l<n.length;l++){var a=n[l],i=a.getSnapshot;a=a.value;try{if(!gl(i(),a))return!1}catch{return!1}}if(n=t.child,t.subtreeFlags&16384&&n!==null)n.return=t,t=n;else{if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return!0;t=t.return}t.sibling.return=t.return,t=t.sibling}}return!0}function Ys(e,t,n,l){t=Gr(e,t),t&=~fu,t&=~za,e.suspendedLanes|=t,e.pingedLanes&=~t,l&&(e.warmLanes|=t),l=e.expirationTimes;for(var a=t;0<a;){var i=31-Gt(a),_=1<<i;l[i]=-1,a&=~_}n!==0&&Fr(e,n,t)}function pu(){return(qt&6)===0?(Yi(0),!1):!0}function jf(){if(Mt!==null){if(Kt===0)var e=Mt.return;else e=Mt,Po=ba=null,Rd(e),_r=null,ji=0,e=Mt;for(;e!==null;)Nh(e.alternate,e),e=e.return;Mt=null}}function Nr(e,t){var n=e.timeoutHandle;return n!==-1&&(e.timeoutHandle=-1,cx(n)),n=e.cancelPendingCommit,n!==null&&(e.cancelPendingCommit=null,n()),Lo=0,jf(),sn=e,Mt=n=Ct(e.current,null),Rt=t,Kt=0,Hl=null,$s=!1,vr=sl(e,t),bf=!1,wr=Yl=fu=za=Us=kn=0,El=Ui=null,vf=!1,ss=Gr(e,t),Pe(),n}function sg(e,t){_t=null,Re.H=Kc,t===fr||t===Lc?(t=fm(),Kt=3):t===yd?(t=fm(),Kt=4):Kt=t===Gd?8:t!==null&&typeof t=="object"&&typeof t.then=="function"?6:1,Hl=t,Mt===null&&(kn=1,Zc(e,Zn(t,e.current)))}function ag(){var e=Pn.current;return e===null?!0:(Rt&4194048)===Rt?cl===null:(Rt&62914560)===Rt||(Rt&536870912)!==0?e===cl:!1}function rg(){var e=Re.H;return Re.H=Kc,e===null?Kc:e}function ig(){var e=Re.A;return Re.A=By,e}function yu(){kn=4,$s||(Rt&4194048)!==Rt&&Pn.current!==null||(vr=!0),(Us&134217727)===0&&(za&134217727)===0||sn===null||Ys(sn,Rt,Yl,!1)}function kf(e,t,n){var l=qt;qt|=2;var a=rg(),i=ig();(sn!==e||Rt!==t)&&(hu=null,Nr(e,t)),t=!1;var _=kn;e:do try{if(Kt!==0&&Mt!==null){var y=Mt,M=Hl;switch(Kt){case 8:jf(),_=6;break e;case 3:case 2:case 9:case 6:Pn.current===null&&(t=!0);var L=Kt;if(Kt=0,Hl=null,Mr(e,y,M,L),n&&vr){_=0;break e}break;default:L=Kt,Kt=0,Hl=null,Mr(e,y,M,L)}}Hy(),_=kn;break}catch(X){sg(e,X)}while(!0);return t&&e.shellSuspendCounter++,Po=ba=null,qt=l,Re.H=a,Re.A=i,Mt===null&&(sn=null,Rt=0,Pe()),_}function Hy(){for(;Mt!==null;)cg(Mt)}function Yy(e,t){var n=qt;qt|=2;var l=rg(),a=ig();sn!==e||Rt!==t?(hu=null,mu=Ze()+500,Nr(e,t)):vr=sl(e,t);e:do try{if(Kt!==0&&Mt!==null){t=Mt;var i=Hl;t:switch(Kt){case 1:Kt=0,Hl=null,Mr(e,t,i,1);break;case 2:case 9:if(um(i)){Kt=0,Hl=null,ug(t);break}t=function(){Kt!==2&&Kt!==9||sn!==e||(Kt=7),Bo(e)},i.then(t,t);break e;case 3:Kt=7;break e;case 4:Kt=5;break e;case 7:um(i)?(Kt=0,Hl=null,ug(t)):(Kt=0,Hl=null,Mr(e,t,i,7));break;case 5:var _=null;switch(Mt.tag){case 26:_=Mt.memoizedState;case 5:case 27:var y=Mt;if(_?t0(_):y.stateNode.complete){Kt=0,Hl=null;var M=y.sibling;if(M!==null)Mt=M;else{var L=y.return;L!==null?(Mt=L,xu(L)):Mt=null}break t}}Kt=0,Hl=null,Mr(e,t,i,5);break;case 6:Kt=0,Hl=null,Mr(e,t,i,6);break;case 8:jf(),kn=6;break e;default:throw Error(u(462))}}qy();break}catch(X){sg(e,X)}while(!0);return Po=ba=null,Re.H=l,Re.A=a,qt=n,Mt!==null?0:(sn=null,Rt=0,Pe(),kn)}function qy(){for(;Mt!==null&&!ct();)cg(Mt)}function cg(e){var t=kh(e.alternate,e,ss);e.memoizedProps=e.pendingProps,t===null?xu(e):Mt=t}function ug(e){var t=e,n=t.alternate;switch(t.tag){case 15:case 0:t=yh(n,t,t.pendingProps,t.type,void 0,Rt);break;case 11:t=yh(n,t,t.pendingProps,t.type.render,t.ref,Rt);break;case 5:Rd(t);var l=t;l===Qn&&(jt?(Rc(l),l.tag===5&&l.stateNode!=null&&(rn=l.stateNode)):(Rc(l),jt=!0));default:Nh(n,t),t=Mt=Wt(t,ss),t=kh(n,t,ss)}e.memoizedProps=e.pendingProps,t===null?xu(e):Mt=t}function Mr(e,t,n,l){Po=ba=null,Rd(t),_r=null,ji=0;var a=t.return;try{if(Ey(e,a,t,n,Rt)){kn=1,Zc(e,Zn(n,e.current)),Mt=null;return}}catch(i){if(a!==null)throw Mt=a,i;kn=1,Zc(e,Zn(n,e.current)),Mt=null;return}t.flags&32768?(jt||l===1?e=!0:vr||(Rt&536870912)!==0?e=!1:($s=e=!0,(l===2||l===9||l===3||l===6)&&(l=Pn.current,l!==null&&l.tag===13&&(l.flags|=16384))),dg(t,e)):xu(t)}function xu(e){var t=e;do{if((t.flags&32768)!==0){dg(t,$s);return}e=t.return;var n=Ay(t.alternate,t,ss);if(n!==null){Mt=n;return}if(t=t.sibling,t!==null){Mt=t;return}Mt=t=e}while(t!==null);kn===0&&(kn=5)}function dg(e,t){do{var n=Oy(e.alternate,e);if(n!==null){n.flags&=32767,Mt=n;return}if(n=e.return,n!==null&&(n.flags|=32768,n.subtreeFlags=0,n.deletions=null),!t&&(e=e.sibling,e!==null)){Mt=e;return}Mt=e=n}while(e!==null);kn=6,Mt=null}function fg(e,t,n,l,a,i,_,y,M,L,X,J){e.cancelPendingCommit=null;do bu();while(xn!==0);if((qt&6)!==0)throw Error(u(327));if(t!==null){if(t===e.current)throw Error(u(177));e===sn&&(Mt=sn=null,Rt=0),Aa=t,fo=e,Lo=n,Sf=a,tg=l,Iy(e,t,n,_,y,M,J)}}function Iy(e,t,n,l,a,i,_){var y=t.lanes|t.childLanes;if(wf=y,y|=qe,vl(e,n,y,l,a,i),jr=null,(n&335544064)===n?(kr=py(e),l=10262):(kr=null,l=10256),(t.subtreeFlags&l)!==0||(t.flags&l)!==0?(e.callbackNode=null,e.callbackPriority=0,Wy(pt,function(){return Ef(),null})):(e.callbackNode=null,e.callbackPriority=0),su=!1,l=(t.flags&13878)!==0,(t.subtreeFlags&13878)!==0||l){l=Re.T,Re.T=null,a=Ue.p,Ue.p=2,i=qt,qt|=4;try{Dy(e,t,n)}finally{qt=i,Ue.p=a,Re.T=l}}xn=1,su?Sr=hx(_,e.containerInfo,kr,Cf,Nf,Qy,Mf,Ef,Xy):(Cf(),Nf(),Mf())}function Xy(e){if(xn!==0){var t=fo.onRecoverableError;t(e,{componentStack:null})}}function Qy(){xn===3&&(xn=0,Fh(Aa,fo),xn=4)}function Cf(){if(xn===1){xn=0;var e=fo,t=Aa,n=Lo,l=(t.flags&13878)!==0;if((t.subtreeFlags&13878)!==0||l){l=Re.T,Re.T=null;var a=Ue.p;Ue.p=2;var i=qt;qt|=4;try{Li=iu=!1,Gh(t,e,n),n=Yf;var _=hi(e.containerInfo),y=n.focusedElem,M=n.selectionRange;if(_!==y&&y&&y.ownerDocument&&mi(y.ownerDocument.documentElement,y)){if(M!==null&&rr(y)){var L=M.start,X=M.end;if(X===void 0&&(X=L),"selectionStart"in y)y.selectionStart=L,y.selectionEnd=Math.min(X,y.value.length);else{var J=y.ownerDocument||document,A=J&&J.defaultView||window;if(A.getSelection){var q=A.getSelection(),ve=y.textContent.length,Le=Math.min(M.start,ve),mt=M.end===void 0?Le:Math.min(M.end,ve);!q.extend&&Le>mt&&(_=mt,mt=Le,Le=_);var D=_i(y,Le),R=_i(y,mt);if(D&&R&&(q.rangeCount!==1||q.anchorNode!==D.node||q.anchorOffset!==D.offset||q.focusNode!==R.node||q.focusOffset!==R.offset)){var U=J.createRange();U.setStart(D.node,D.offset),q.removeAllRanges(),Le>mt?(q.addRange(U),q.extend(R.node,R.offset)):(U.setEnd(R.node,R.offset),q.addRange(U))}}}}for(J=[],q=y;q=q.parentNode;)q.nodeType===1&&J.push({element:q,left:q.scrollLeft,top:q.scrollTop});for(typeof y.focus=="function"&&y.focus(),y=0;y<J.length;y++){var Z=J[y];Z.element.scrollLeft=Z.left,Z.element.scrollTop=Z.top}}Lr=!!Hf,Yf=Hf=null}finally{qt=i,Ue.p=a,Re.T=l}}e.current=t,xn=2}}function Nf(){if(xn===2){xn=0;var e=fo,t=Aa,n=(t.flags&8772)!==0;if((t.subtreeFlags&8772)!==0||n){n=Re.T,Re.T=null;var l=Ue.p;Ue.p=2;var a=qt;qt|=4;try{Hh(e,t.alternate,t)}finally{qt=a,Ue.p=l,Re.T=n}}xn=3}}function Mf(){if(xn===4||xn===3){xn=0;var e=Sr;Sr=null,Ke();var t=fo,n=Aa,l=Lo,a=tg,i=(l&335544064)===l?10262:10256;if((n.subtreeFlags&i)!==0||(n.flags&i)!==0?xn=5:(xn=0,Aa=fo=null,_g(t,t.pendingLanes)),i=t.pendingLanes,i===0&&(Hs=null),$a(l),n=n.stateNode,Je&&typeof Je.onCommitFiberRoot=="function")try{Je.onCommitFiberRoot(vt,n,void 0,(n.current.flags&128)===128)}catch{}if(a!==null){n=Re.T,i=Ue.p,Ue.p=2,Re.T=null;try{for(var _=t.onRecoverableError,y=0;y<a.length;y++){var M=a[y];_(M.value,{componentStack:M.stack})}}finally{Re.T=n,Ue.p=i}}if(a=jr,_=kr,kr=null,a!==null&&(jr=null,_===null&&(_=[]),e!==null))for(M=0;M<a.length;M++)n=(0,a[M])(_),n!==void 0&&e.finished.finally(n);(Lo&3)!==0&&bu(),Bo(t),i=t.pendingLanes,(l&261930)!==0&&(i&42)!==0?t===gu?Hi++:(Hi=0,gu=t):(Hi=0,gu=null),Yi(0)}}function _g(e,t){(e.pooledCacheLanes&=t)===0&&(t=e.pooledCache,t!=null&&(e.pooledCache=null,vi(t)))}function bu(){return Sr!==null&&(Sr.skipTransition(),Sr=null),Cf(),Nf(),Mf(),Ef()}function Ef(){if(xn!==5)return!1;var e=fo,t=wf;wf=0;var n=$a(Lo),l=Re.T,a=Ue.p;try{Ue.p=32>n?32:n,Re.T=null,n=Sf,Sf=null;var i=fo,_=Lo;if(xn=0,Aa=fo=null,Lo=0,(qt&6)!==0)throw Error(u(331));var y=qt;if(qt|=4,Jh(i.current),Wh(i,i.current,_,n),qt=y,Yi(0,!1),Je&&typeof Je.onPostCommitFiberRoot=="function")try{Je.onPostCommitFiberRoot(vt,i)}catch{}return!0}finally{Ue.p=a,Re.T=l,_g(e,t)}}function mg(e,t,n){t=Zn(n,t),t=Qd(e.stateNode,t,2),e=zs(e,t,2),e!==null&&(Ps(e,2),Bo(e))}function Zt(e,t,n){if(e.tag===3)mg(e,e,n);else for(;t!==null;){if(t.tag===3){mg(t,e,n);break}else if(t.tag===1){var l=t.stateNode;if(typeof t.type.getDerivedStateFromError=="function"||typeof l.componentDidCatch=="function"&&(Hs===null||!Hs.has(l))){e=Zn(n,e),n=uh(2),l=zs(t,n,2),l!==null&&(dh(n,l,t,e),Ps(l,2),Bo(l));break}}t=t.return}}function Tf(e,t,n){var l=e.pingCache;if(l===null){l=e.pingCache=new $y;var a=new Set;l.set(t,a)}else a=l.get(t),a===void 0&&(a=new Set,l.set(t,a));a.has(n)||(bf=!0,a.add(n),e=Gy.bind(null,e,t,n),t.then(e,e))}function Gy(e,t,n){var l=e.pingCache;l!==null&&l.delete(t),e.pingedLanes|=e.suspendedLanes&n,e.warmLanes&=~n,sn===e&&(Rt&n)===n&&((kn===4||kn===3&&(Rt&62914560)===Rt&&300>Ze()-_u)&&(qt&2)===0?Nr(e,0):fu|=n,wr===Rt&&(wr=0)),Bo(e)}function hg(e,t){t===0&&(t=Vr()),e=Ve(e,t),e!==null&&(Ps(e,t),Bo(e))}function Vy(e){var t=e.memoizedState,n=0;t!==null&&(n=t.retryLane),hg(e,n)}function Fy(e,t){var n=0;switch(e.tag){case 31:case 13:var l=e.stateNode,a=e.memoizedState;a!==null&&(n=a.retryLane);break;case 19:l=e.stateNode;break;case 22:l=e.stateNode._retryCache;break;default:throw Error(u(314))}l!==null&&l.delete(t),hg(e,n)}function Wy(e,t){return at(e,t)}var Er=null,Tr=null,Rf=!1,vu=!1,zf=!1,qs=0;function Bo(e){e!==Tr&&e.next===null&&(Tr===null?Er=Tr=e:Tr=Tr.next=e),vu=!0,Rf||(Rf=!0,Zy())}function Yi(e,t){if(!zf&&vu){zf=!0;do for(var n=!1,l=Er;l!==null;){if(e!==0){var a=l.pendingLanes;if(a===0)var i=0;else{var _=l.suspendedLanes,y=l.pingedLanes;i=(1<<31-Gt(42|e)+1)-1,i&=a&~(_&~y),i=i&201326741?i&201326741|1:i?i|2:0}i!==0&&(n=!0,xg(l,i))}else i=Rt,i=dl(l,l===sn?i:0,l.cancelPendingCommit!==null||l.timeoutHandle!==-1),(i&3)===0||sl(l,i)||(n=!0,xg(l,i));l=l.next}while(n);zf=!1}}function Ky(){gg()}function gg(){vu=Rf=!1;var e=0;qs!==0&&ix()&&(e=qs);for(var t=Ze(),n=null,l=Er;l!==null;){var a=l.next,i=pg(l,t);i===0?(l.next=null,n===null?Er=a:n.next=a,a===null&&(Tr=n)):(n=l,(e!==0||(i&3)!==0)&&(vu=!0)),l=a}xn!==0&&xn!==5||Yi(e),qs!==0&&(qs=0)}function pg(e,t){for(var n=e.suspendedLanes,l=e.pingedLanes,a=e.expirationTimes,i=e.pendingLanes&-62914561;0<i;){var _=31-Gt(i),y=1<<_,M=a[_];M===-1?((y&n)===0||(y&l)!==0)&&(a[_]=hc(y,t)):M<=t&&(e.expiredLanes|=y),i&=~y}if(t=sn,n=Rt,n=dl(e,e===t?n:0,e.cancelPendingCommit!==null||e.timeoutHandle!==-1),l=e.callbackNode,n===0||e===t&&(Kt===2||Kt===9)||e.cancelPendingCommit!==null)return l!==null&&l!==null&&ot(l),e.callbackNode=null,e.callbackPriority=0;if((n&3)===0||sl(e,n)){if(t=n&-n,t===e.callbackPriority)return t;switch(l!==null&&ot(l),$a(n)){case 2:case 8:n=pn;break;case 32:n=pt;break;case 268435456:n=kt;break;default:n=pt}return l=yg.bind(null,e),n=at(n,l),e.callbackPriority=t,e.callbackNode=n,t}return l!==null&&l!==null&&ot(l),e.callbackPriority=2,e.callbackNode=null,2}function yg(e,t){if(xn!==0&&xn!==5)return e.callbackNode=null,e.callbackPriority=0,null;var n=e.callbackNode;if(bu()&&e.callbackNode!==n)return null;var l=Rt;return l=dl(e,e===sn?l:0,e.cancelPendingCommit!==null||e.timeoutHandle!==-1),l===0?null:(lg(e,l,t),pg(e,Ze()),e.callbackNode!=null&&e.callbackNode===n?yg.bind(null,e):null)}function xg(e,t){if(bu())return null;lg(e,t,!0)}function Zy(){ux(function(){(qt&6)!==0?at(Bt,Ky):gg()})}function Af(){if(qs===0){var e=Sa;e===0&&(e=yn,yn<<=1,(yn&261888)===0&&(yn=256)),qs=e}return qs}function bg(e){return e==null||typeof e=="symbol"||typeof e=="boolean"?null:typeof e=="function"?e:la(e)}function Jy(e,t,n,l,a){if(t==="submit"&&n&&n.stateNode===a){var i=bg((a[wn]||null).action),_=l.submitter;_&&(t=(t=_[wn]||null)?bg(t.formAction):_.getAttribute("formAction"),t!==null&&(i=t,_=null));var y=new Sl("action","action",null,l,a);e.push({event:y,listeners:[{instance:null,listener:function(){if(l.defaultPrevented){if(qs!==0){var M=new FormData(a,_);Hd(n,{pending:!0,data:M,method:a.method,action:i},null,M)}}else typeof i=="function"&&(y.preventDefault(),M=new FormData(a,_),Hd(n,{pending:!0,data:M,method:a.method,action:i},i,M))},currentTarget:a}]})}}for(var Of=0;Of<N.length;Of++){var Df=N[Of],Py=Df.toLowerCase(),ex=Df[0].toUpperCase()+Df.slice(1);T(Py,"on"+ex)}T(pi,"onAnimationEnd"),T(yi,"onAnimationIteration"),T(xi,"onAnimationStart"),T("dblclick","onDoubleClick"),T("focusin","onFocus"),T("focusout","onBlur"),T(Nc,"onTransitionRun"),T(Mc,"onTransitionStart"),T(Ec,"onTransitionCancel"),T(x,"onTransitionEnd"),ds("onMouseEnter",["mouseout","mouseover"]),ds("onMouseLeave",["mouseout","mouseover"]),ds("onPointerEnter",["pointerout","pointerover"]),ds("onPointerLeave",["pointerout","pointerover"]),Ho("onChange","change click focusin focusout input keydown keyup selectionchange".split(" ")),Ho("onSelect","focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" ")),Ho("onBeforeInput",["compositionend","keypress","textInput","paste"]),Ho("onCompositionEnd","compositionend focusout keydown keypress keyup mousedown".split(" ")),Ho("onCompositionStart","compositionstart focusout keydown keypress keyup mousedown".split(" ")),Ho("onCompositionUpdate","compositionupdate focusout keydown keypress keyup mousedown".split(" "));var qi="abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "),tx=new Set("beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(qi));function vg(e,t){t=(t&4)!==0;for(var n=0;n<e.length;n++){var l=e[n],a=l.event;l=l.listeners;e:{var i=void 0;if(t)for(var _=l.length-1;0<=_;_--){var y=l[_],M=y.instance,L=y.currentTarget;if(y=y.listener,M!==i&&a.isPropagationStopped())break e;i=y,a.currentTarget=L;try{i(a)}catch(X){me(X)}a.currentTarget=null,i=M}else for(_=0;_<l.length;_++){if(y=l[_],M=y.instance,L=y.currentTarget,y=y.listener,M!==i&&a.isPropagationStopped())break e;i=y,a.currentTarget=L;try{i(a)}catch(X){me(X)}a.currentTarget=null,i=M}}}}function Et(e,t){var n=t[ft];n===void 0&&(n=t[ft]=new Set);var l=e+"__bubble";n.has(l)||(wg(t,e,2,!1),n.add(l))}function Lf(e,t,n){var l=0;t&&(l|=4),wg(n,e,l,t)}var wu="_reactListening"+Math.random().toString(36).slice(2);function Bf(e){if(!e[wu]){e[wu]=!0,qn.forEach(function(n){n!=="selectionchange"&&(tx.has(n)||Lf(n,!1,e),Lf(n,!0,e))});var t=e.nodeType===9?e:e.ownerDocument;t===null||t[wu]||(t[wu]=!0,Lf("selectionchange",!1,t))}}function wg(e,t,n,l){switch(d0(t)){case 2:var a=Vx;break;case 8:a=Fx;break;default:a=l_}n=a.bind(null,t,n,e),a=void 0,!Ga||t!=="touchstart"&&t!=="touchmove"&&t!=="wheel"||(a=!0),l?a!==void 0?e.addEventListener(t,n,{capture:!0,passive:a}):e.addEventListener(t,n,!0):a!==void 0?e.addEventListener(t,n,{passive:a}):e.addEventListener(t,n,!1)}function $f(e,t,n,l,a){var i=l;if((t&1)===0&&(t&2)===0&&l!==null)e:for(;;){if(l===null)return;var _=l.tag;if(_===3||_===4){var y=l.stateNode.containerInfo;if(y===a)break;if(_===4)for(_=l.return;_!==null;){var M=_.tag;if((M===3||M===4)&&_.stateNode.containerInfo===a)return;_=_.return}for(;y!==null;){if(_=yo(y),_===null)return;if(M=_.tag,M===5||M===6||M===26||M===27){l=i=_;continue e}y=y.parentNode}}l=l.return}hs(function(){var L=i,X=sa(n),J=[];e:{var A=S.get(e);if(A!==void 0){var q=Sl,ve=e;switch(e){case"keypress":if(aa(n)===0)break e;case"keydown":case"keyup":q=jl;break;case"focusin":ve="focus",q=ps;break;case"focusout":ve="blur",q=ps;break;case"beforeblur":case"afterblur":q=ps;break;case"click":if(n.button===2)break e;case"auxclick":case"dblclick":case"mousedown":case"mousemove":case"mouseup":case"mouseout":case"mouseover":case"contextmenu":q=Fa;break;case"drag":case"dragend":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"dragstart":case"drop":q=yc;break;case"touchcancel":case"touchend":case"touchmove":case"touchstart":q=vc;break;case pi:case yi:case xi:q=Ka;break;case x:q=id;break;case"scroll":case"scrollend":q=li;break;case"wheel":q=Ja;break;case"copy":case"cut":case"paste":q=So;break;case"gotpointercapture":case"lostpointercapture":case"pointercancel":case"pointerdown":case"pointermove":case"pointerout":case"pointerover":case"pointerup":q=Vo;break;case"submit":q=bc;break;case"toggle":case"beforetoggle":q=cd}var Le=(t&4)!==0,mt=!Le&&(e==="scroll"||e==="scrollend"),D=Le?A!==null?A+"Capture":null:A;Le=[];for(var R=L,U;R!==null;){var Z=R;if(U=Z.stateNode,Z=Z.tag,Z!==5&&Z!==26&&Z!==27||U===null||D===null||(Z=Sn(R,D),Z!=null&&Le.push(Ii(R,Z,U))),mt)break;R=R.return}0<Le.length&&(A=new q(A,ve,null,n,X),J.push({event:A,listeners:Le}))}}if((t&7)===0){e:{if(q=e==="mouseover"||e==="pointerover",A=e==="mouseout"||e==="pointerout",q&&n!==oa&&(ve=n.relatedTarget||n.fromElement)&&(yo(ve)||ve[Gl]))break e;(A||q)&&(ve=X.window===X?X:(q=X.ownerDocument)?q.defaultView||q.parentWindow:window,A?(q=n.relatedTarget||n.toElement,A=L,q=q?yo(q):null,q!==null&&(mt=d(q),Le=q.tag,q!==mt||Le!==5&&Le!==27&&Le!==6)&&(q=null)):(A=null,q=L),A!==q&&(Le=Fa,Z="onMouseLeave",D="onMouseEnter",R="mouse",(e==="pointerout"||e==="pointerover")&&(Le=Vo,Z="onPointerLeave",D="onPointerEnter",R="pointer"),mt=A==null?ve:cs(A),U=q==null?ve:cs(q),ve=new Le(Z,R+"leave",A,n,X),ve.target=mt,ve.relatedTarget=U,Z=null,yo(X)===L&&(Le=new Le(D,R+"enter",q,n,X),Le.target=U,Le.relatedTarget=mt,Z=Le),mt=Z,Le=A&&q?de(A,q,nx):null,A!==null&&Sg(J,ve,A,Le,!1),q!==null&&mt!==null&&Sg(J,mt,q,Le,!0)))}e:{if(A=L?cs(L):window,q=A.nodeName&&A.nodeName.toLowerCase(),q==="select"||q==="input"&&A.type==="file")var ze=sr;else if(bs(A))if(ga)ze=ar;else{ze=di;var zt=ui}else q=A.nodeName,!q||q.toLowerCase()!=="input"||A.type!=="checkbox"&&A.type!=="radio"?L&&_l(L.elementType)&&(ze=sr):ze=Fl;if(ze&&(ze=ze(e,L))){ma(J,ze,n,X);break e}zt&&zt(e,A,L)}switch(zt=L?cs(L):window,e){case"focusin":(bs(zt)||zt.contentEditable==="true")&&(Fo=zt,rl=L,Ss=null);break;case"focusout":Ss=rl=Fo=null;break;case"mousedown":ir=!0;break;case"contextmenu":case"mouseup":case"dragend":ir=!1,gi(J,n,X);break;case"selectionchange":if(Cc)break;case"keydown":case"keyup":gi(J,n,X)}var Xe;if(er)e:{switch(e){case"compositionstart":var tt="onCompositionStart";break e;case"compositionend":tt="onCompositionEnd";break e;case"compositionupdate":tt="onCompositionUpdate";break e}tt=void 0}else ao?lr(e,n)&&(tt="onCompositionEnd"):e==="keydown"&&n.keyCode===229&&(tt="onCompositionStart");tt&&(tr&&n.locale!=="ko"&&(ao||tt!=="onCompositionStart"?tt==="onCompositionEnd"&&ao&&(Xe=ni()):(wt=X,ti="value"in wt?wt.value:wt.textContent,ao=!0)),zt=Su(L,tt),0<zt.length&&(tt=new hl(tt,e,null,n,X),J.push({event:tt,listeners:zt}),Xe?tt.data=Xe:(Xe=xs(n),Xe!==null&&(tt.data=Xe)))),(Xe=wc?jc(e,n):or(e,n))&&(tt=Su(L,"onBeforeInput"),0<tt.length&&(zt=new hl("onBeforeInput","beforeinput",null,n,X),J.push({event:zt,listeners:tt}),zt.data=Xe)),Jy(J,e,L,n,X)}vg(J,t)})}function Ii(e,t,n){return{instance:e,listener:t,currentTarget:n}}function Su(e,t){for(var n=t+"Capture",l=[];e!==null;){var a=e,i=a.stateNode;if(a=a.tag,a!==5&&a!==26&&a!==27||i===null||(a=Sn(e,n),a!=null&&l.unshift(Ii(e,a,i)),a=Sn(e,t),a!=null&&l.push(Ii(e,a,i))),e.tag===3)return l;e=e.return}return[]}function nx(e){if(e===null)return null;do e=e.return;while(e&&e.tag!==5&&e.tag!==27);return e||null}function Sg(e,t,n,l,a){for(var i=t._reactName,_=[];n!==null&&n!==l;){var y=n,M=y.alternate,L=y.stateNode;if(y=y.tag,M!==null&&M===l)break;y!==5&&y!==26&&y!==27||L===null||(M=L,a?(L=Sn(n,i),L!=null&&_.unshift(Ii(n,L,M))):a||(L=Sn(n,i),L!=null&&_.push(Ii(n,L,M)))),n=n.return}_.length!==0&&e.push({event:t,listeners:_})}var lx=/\r\n?/g,ox=/\u0000|\uFFFD/g;function jg(e){return(typeof e=="string"?e:""+e).replace(lx,`
`).replace(ox,"")}function kg(e,t){return t=jg(t),jg(e)===t}function Jt(e,t,n,l,a,i){switch(n){case"children":if(typeof l=="string")t==="body"||t==="textarea"&&l===""||vo(e,l);else if(typeof l=="number"||typeof l=="bigint")t!=="body"&&vo(e,""+l);else return;break;case"className":Ya(e,"class",l);break;case"tabIndex":Ya(e,"tabindex",l);break;case"dir":case"role":case"viewBox":case"width":case"height":Ya(e,n,l);break;case"style":Qa(e,l,i);return;case"data":if(t!=="object"){Ya(e,"data",l);break}case"src":case"href":if(l===""&&(t!=="a"||n!=="href")){e.removeAttribute(n);break}if(l==null||typeof l=="function"||typeof l=="symbol"||typeof l=="boolean"){e.removeAttribute(n);break}l=la(l),e.setAttribute(n,l);break;case"action":case"formAction":if(typeof l=="function"){e.setAttribute(n,"javascript:throw new Error('A React form was unexpectedly submitted. If you called form.submit() manually, consider using form.requestSubmit() instead. If you\\'re trying to use event.stopPropagation() in a submit event handler, consider also calling event.preventDefault().')");break}else typeof i=="function"&&(n==="formAction"?(t!=="input"&&Jt(e,t,"name",a.name,a,null),Jt(e,t,"formEncType",a.formEncType,a,null),Jt(e,t,"formMethod",a.formMethod,a,null),Jt(e,t,"formTarget",a.formTarget,a,null)):(Jt(e,t,"encType",a.encType,a,null),Jt(e,t,"method",a.method,a,null),Jt(e,t,"target",a.target,a,null)));if(l==null||typeof l=="symbol"||typeof l=="boolean"){e.removeAttribute(n);break}l=la(l),e.setAttribute(n,l);break;case"onClick":l!=null&&(e.onclick=no);return;case"onScroll":l!=null&&Et("scroll",e);return;case"onScrollEnd":l!=null&&Et("scrollend",e);return;case"dangerouslySetInnerHTML":if(l!=null){if(typeof l!="object"||!("__html"in l))throw Error(u(61));if(n=l.__html,n!=null){if(a.children!=null)throw Error(u(60));(i!=null?i.__html:void 0)!==n&&(e.innerHTML=n)}}break;case"multiple":e.multiple=l&&typeof l!="function"&&typeof l!="symbol";break;case"muted":e.muted=l&&typeof l!="function"&&typeof l!="symbol";break;case"suppressContentEditableWarning":case"suppressHydrationWarning":case"defaultValue":case"defaultChecked":case"innerHTML":case"ref":break;case"autoFocus":break;case"xlinkHref":if(l==null||typeof l=="function"||typeof l=="boolean"||typeof l=="symbol"){e.removeAttribute("xlink:href");break}n=la(l),e.setAttributeNS("http://www.w3.org/1999/xlink","xlink:href",n);break;case"contentEditable":case"spellCheck":case"draggable":case"value":case"autoReverse":case"externalResourcesRequired":case"focusable":case"preserveAlpha":l!=null&&typeof l!="function"&&typeof l!="symbol"?e.setAttribute(n,l):e.removeAttribute(n);break;case"inert":case"allowFullScreen":case"async":case"autoPlay":case"controls":case"credentialless":case"default":case"defer":case"disabled":case"disablePictureInPicture":case"disableRemotePlayback":case"formNoValidate":case"hidden":case"loop":case"noModule":case"noValidate":case"open":case"playsInline":case"readOnly":case"required":case"reversed":case"scoped":case"seamless":case"itemScope":l&&typeof l!="function"&&typeof l!="symbol"?e.setAttribute(n,""):e.removeAttribute(n);break;case"capture":case"download":l===!0?e.setAttribute(n,""):l!==!1&&l!=null&&typeof l!="function"&&typeof l!="symbol"?e.setAttribute(n,l):e.removeAttribute(n);break;case"cols":case"rows":case"size":case"span":l!=null&&typeof l!="function"&&typeof l!="symbol"&&!isNaN(l)&&1<=l?e.setAttribute(n,l):e.removeAttribute(n);break;case"rowSpan":case"start":l==null||typeof l=="function"||typeof l=="symbol"||isNaN(l)?e.removeAttribute(n):e.setAttribute(n,l);break;case"popover":Et("beforetoggle",e),Et("toggle",e),na(e,"popover",l);break;case"xlinkActuate":to(e,"http://www.w3.org/1999/xlink","xlink:actuate",l);break;case"xlinkArcrole":to(e,"http://www.w3.org/1999/xlink","xlink:arcrole",l);break;case"xlinkRole":to(e,"http://www.w3.org/1999/xlink","xlink:role",l);break;case"xlinkShow":to(e,"http://www.w3.org/1999/xlink","xlink:show",l);break;case"xlinkTitle":to(e,"http://www.w3.org/1999/xlink","xlink:title",l);break;case"xlinkType":to(e,"http://www.w3.org/1999/xlink","xlink:type",l);break;case"xmlBase":to(e,"http://www.w3.org/XML/1998/namespace","xml:base",l);break;case"xmlLang":to(e,"http://www.w3.org/XML/1998/namespace","xml:lang",l);break;case"xmlSpace":to(e,"http://www.w3.org/XML/1998/namespace","xml:space",l);break;case"is":na(e,"is",l);break;case"innerText":case"textContent":return;default:if(!(2<n.length)||n[0]!=="o"&&n[0]!=="O"||n[1]!=="n"&&n[1]!=="N")n=rd.get(n)||n,na(e,n,l);else return}yt=!0}function Uf(e,t,n,l,a,i){switch(n){case"style":Qa(e,l,i);return;case"dangerouslySetInnerHTML":if(l!=null){if(typeof l!="object"||!("__html"in l))throw Error(u(61));if(n=l.__html,n!=null){if(a.children!=null)throw Error(u(60));(i!=null?i.__html:void 0)!==n&&(e.innerHTML=n)}}break;case"children":if(typeof l=="string")vo(e,l);else if(typeof l=="number"||typeof l=="bigint")vo(e,""+l);else return;break;case"onScroll":l!=null&&Et("scroll",e);return;case"onScrollEnd":l!=null&&Et("scrollend",e);return;case"onClick":l!=null&&(e.onclick=no);return;case"suppressContentEditableWarning":case"suppressHydrationWarning":case"innerHTML":case"ref":return;case"innerText":case"textContent":return;default:if(!ta.hasOwnProperty(n))e:{if(n[0]==="o"&&n[1]==="n"&&(a=n.endsWith("Capture"),i=n.slice(2,a?n.length-7:void 0),t=e[wn]||null,t=t!=null?t[n]:null,typeof t=="function"&&e.removeEventListener(i,t,a),typeof l=="function")){typeof t!="function"&&t!==null&&(n in e?e[n]=null:e.hasAttribute(n)&&e.removeAttribute(n)),e.addEventListener(i,l,a);break e}yt=!0,n in e?e[n]=l:l===!0?e.setAttribute(n,""):na(e,n,l)}return}yt=!0}function nl(e,t,n){switch(t){case"div":case"span":case"svg":case"path":case"a":case"g":case"p":case"li":break;case"img":Et("error",e),Et("load",e);var l=!1,a=!1,i;for(i in n)if(n.hasOwnProperty(i)){var _=n[i];if(_!=null)switch(i){case"src":l=!0;break;case"srcSet":a=!0;break;case"children":case"dangerouslySetInnerHTML":throw Error(u(137,t));default:Jt(e,t,i,_,n,null)}}a&&Jt(e,t,"srcSet",n.srcSet,n,null),l&&Jt(e,t,"src",n.src,n,null);return;case"input":Et("invalid",e);var y=i=_=a=null,M=null,L=null;for(l in n)if(n.hasOwnProperty(l)){var X=n[l];if(X!=null)switch(l){case"name":a=X;break;case"type":_=X;break;case"checked":M=X;break;case"defaultChecked":L=X;break;case"value":i=X;break;case"defaultValue":y=X;break;case"children":case"dangerouslySetInnerHTML":if(X!=null)throw Error(u(137,t));break;default:Jt(e,t,l,X,n,null)}}wl(e,i,y,M,L,_,a,!1);return;case"select":Et("invalid",e),l=_=i=null;for(a in n)if(n.hasOwnProperty(a)&&(y=n[a],y!=null))switch(a){case"value":i=y;break;case"defaultValue":_=y;break;case"multiple":l=y;default:Jt(e,t,a,y,n,null)}t=i,n=_,e.multiple=!!l,t!=null?Yo(e,!!l,t,!1):n!=null&&Yo(e,!!l,n,!0);return;case"textarea":Et("invalid",e),i=a=l=null;for(_ in n)if(n.hasOwnProperty(_)&&(y=n[_],y!=null))switch(_){case"value":l=y;break;case"defaultValue":a=y;break;case"children":i=y;break;case"dangerouslySetInnerHTML":if(y!=null)throw Error(u(91));break;default:Jt(e,t,_,y,n,null)}Xa(e,l,a,i);return;case"option":for(M in n)if(n.hasOwnProperty(M)&&(l=n[M],l!=null))switch(M){case"selected":e.selected=l&&typeof l!="function"&&typeof l!="symbol";break;default:Jt(e,t,M,l,n,null)}return;case"dialog":Et("beforetoggle",e),Et("toggle",e),Et("cancel",e),Et("close",e);break;case"iframe":case"object":Et("load",e);break;case"video":case"audio":for(l=0;l<qi.length;l++)Et(qi[l],e);break;case"image":Et("error",e),Et("load",e);break;case"details":Et("toggle",e);break;case"embed":case"source":case"link":Et("error",e),Et("load",e);case"area":case"base":case"br":case"col":case"hr":case"keygen":case"meta":case"param":case"track":case"wbr":case"menuitem":for(L in n)if(n.hasOwnProperty(L)&&(l=n[L],l!=null))switch(L){case"children":case"dangerouslySetInnerHTML":throw Error(u(137,t));default:Jt(e,t,L,l,n,null)}return;default:if(_l(t)){for(X in n)n.hasOwnProperty(X)&&(l=n[X],l!==void 0&&Uf(e,t,X,l,n,void 0));return}}for(y in n)n.hasOwnProperty(y)&&(l=n[y],l!=null&&Jt(e,t,y,l,n,null))}var sx={};function ax(e,t,n,l){switch(t){case"div":case"span":case"svg":case"path":case"a":case"g":case"p":case"li":break;case"input":var a=null,i=null,_=null,y=null,M=null,L=null,X=null;for(q in n){var J=n[q];if(n.hasOwnProperty(q)&&J!=null)switch(q){case"checked":break;case"value":break;case"defaultValue":M=J;default:l.hasOwnProperty(q)||Jt(e,t,q,null,l,J)}}for(var A in l){var q=l[A];if(J=n[A],l.hasOwnProperty(A)&&(q!=null||J!=null))switch(A){case"type":q!==J&&(yt=!0),i=q;break;case"name":q!==J&&(yt=!0),a=q;break;case"checked":q!==J&&(yt=!0),L=q;break;case"defaultChecked":q!==J&&(yt=!0),X=q;break;case"value":q!==J&&(yt=!0),_=q;break;case"defaultValue":q!==J&&(yt=!0),y=q;break;case"children":case"dangerouslySetInnerHTML":if(q!=null)throw Error(u(137,t));break;default:q!==J&&Jt(e,t,A,q,l,J)}}ms(e,_,y,M,L,X,i,a);return;case"select":q=_=y=A=null;for(i in n)if(M=n[i],n.hasOwnProperty(i)&&M!=null)switch(i){case"value":break;case"multiple":q=M;default:l.hasOwnProperty(i)||Jt(e,t,i,null,l,M)}for(a in l)if(i=l[a],M=n[a],l.hasOwnProperty(a)&&(i!=null||M!=null))switch(a){case"value":i!==M&&(yt=!0),A=i;break;case"defaultValue":i!==M&&(yt=!0),y=i;break;case"multiple":i!==M&&(yt=!0),_=i;default:i!==M&&Jt(e,t,a,i,l,M)}t=y,n=_,l=q,A!=null?Yo(e,!!n,A,!1):!!l!=!!n&&(t!=null?Yo(e,!!n,t,!0):Yo(e,!!n,n?[]:"",!1));return;case"textarea":q=A=null;for(y in n)if(a=n[y],n.hasOwnProperty(y)&&a!=null&&!l.hasOwnProperty(y))switch(y){case"value":break;case"children":break;default:Jt(e,t,y,null,l,a)}for(_ in l)if(a=l[_],i=n[_],l.hasOwnProperty(_)&&(a!=null||i!=null))switch(_){case"value":a!==i&&(yt=!0),A=a;break;case"defaultValue":a!==i&&(yt=!0),q=a;break;case"children":break;case"dangerouslySetInnerHTML":if(a!=null)throw Error(u(91));break;default:a!==i&&Jt(e,t,_,a,l,i)}Ia(e,A,q);return;case"option":for(var ve in n)if(A=n[ve],n.hasOwnProperty(ve)&&A!=null&&!l.hasOwnProperty(ve))switch(ve){case"selected":e.selected=!1;break;default:Jt(e,t,ve,null,l,A)}for(M in l)if(A=l[M],q=n[M],l.hasOwnProperty(M)&&A!==q&&(A!=null||q!=null))switch(M){case"selected":A!==q&&(yt=!0),e.selected=A&&typeof A!="function"&&typeof A!="symbol";break;default:Jt(e,t,M,A,l,q)}return;case"img":case"link":case"area":case"base":case"br":case"col":case"embed":case"hr":case"keygen":case"meta":case"param":case"source":case"track":case"wbr":case"menuitem":for(var Le in n)A=n[Le],n.hasOwnProperty(Le)&&A!=null&&!l.hasOwnProperty(Le)&&Jt(e,t,Le,null,l,A);for(L in l)if(A=l[L],q=n[L],l.hasOwnProperty(L)&&A!==q&&(A!=null||q!=null))switch(L){case"children":case"dangerouslySetInnerHTML":if(A!=null)throw Error(u(137,t));break;default:Jt(e,t,L,A,l,q)}return;default:if(_l(t)){for(var mt in n)A=n[mt],n.hasOwnProperty(mt)&&A!==void 0&&!l.hasOwnProperty(mt)&&Uf(e,t,mt,void 0,l,A);for(X in l)A=l[X],q=n[X],!l.hasOwnProperty(X)||A===q||A===void 0&&q===void 0||Uf(e,t,X,A,l,q);return}}for(var D in n)A=n[D],n.hasOwnProperty(D)&&A!=null&&!l.hasOwnProperty(D)&&Jt(e,t,D,null,l,A);for(J in l)A=l[J],q=n[J],!l.hasOwnProperty(J)||A===q||A==null&&q==null||Jt(e,t,J,A,l,q)}function Cg(e){switch(e){case"css":case"script":case"font":case"img":case"image":case"input":case"link":return!0;default:return!1}}function rx(){if(typeof performance.getEntriesByType=="function"){for(var e=0,t=0,n=performance.getEntriesByType("resource"),l=0;l<n.length;l++){var a=n[l],i=a.transferSize,_=a.initiatorType,y=a.duration;if(i&&y&&Cg(_)){for(_=0,y=a.responseEnd,l+=1;l<n.length;l++){var M=n[l],L=M.startTime;if(L>y)break;var X=M.transferSize,J=M.initiatorType;X&&Cg(J)&&(M=M.responseEnd,_+=X*(M<y?1:(y-L)/(M-L)))}if(--l,t+=8*(i+_)/(a.duration/1e3),e++,10<e)break}}if(0<e)return t/e/1e6}return navigator.connection&&(e=navigator.connection.downlink,typeof e=="number")?e:5}var Hf=null,Yf=null;function Xi(e){return e.nodeType===9?e:e.ownerDocument}function Ng(e){switch(e){case"http://www.w3.org/2000/svg":return 1;case"http://www.w3.org/1998/Math/MathML":return 2;default:return 0}}function Mg(e,t){if(e===0)switch(t){case"svg":return 1;case"math":return 2;default:return 0}return e===1&&t==="foreignObject"?0:e}function Eg(e,t,n,l){return n=Xi(n).createElement(e),n[En]=l,n[wn]=t,nl(n,e,t),Tn(n),n}function qf(e,t){return e==="textarea"||e==="noscript"||typeof t.children=="string"||typeof t.children=="number"||typeof t.children=="bigint"||typeof t.dangerouslySetInnerHTML=="object"&&t.dangerouslySetInnerHTML!==null&&t.dangerouslySetInnerHTML.__html!=null}var If=null;function ix(){var e=window.event;return e&&e.type==="popstate"?e===If?!1:(If=e,!0):(If=null,!1)}var Xf=typeof setTimeout=="function"?setTimeout:void 0,cx=typeof clearTimeout=="function"?clearTimeout:void 0,Tg=typeof Promise=="function"?Promise:void 0,Rg=typeof requestAnimationFrame=="function"?requestAnimationFrame:Xf,ux=typeof queueMicrotask=="function"?queueMicrotask:typeof Tg<"u"?function(e){return Tg.resolve(null).then(e).catch(dx)}:Xf;function dx(e){setTimeout(function(){throw e})}function Is(e){return e==="head"}function zg(e,t){var n=t,l=0;do{var a=n.nextSibling;if(e.removeChild(n),a&&a.nodeType===8)if(n=a.data,n==="/$"||n==="/&"){if(l===0){e.removeChild(a),Br(t);return}l--}else if(n==="$"||n==="$?"||n==="$~"||n==="$!"||n==="&")l++;else if(n==="html")Jf(e.ownerDocument.documentElement);else if(n==="head"){n=e.ownerDocument.head,Jf(n);for(var i=n.firstChild;i;){var _=i.nextSibling,y=i.nodeName;i[Uo]||y==="SCRIPT"||y==="STYLE"||y==="LINK"&&i.rel.toLowerCase()==="stylesheet"||n.removeChild(i),i=_}}else n==="body"&&Jf(e.ownerDocument.body);n=a}while(n);Br(t)}function Ag(e,t){var n=e;e=0;do{var l=n.nextSibling;if(n.nodeType===1?t?(n._stashedDisplay=n.style.display,n.style.display="none"):(n.style.display=n._stashedDisplay||"",n.getAttribute("style")===""&&n.removeAttribute("style")):n.nodeType===3&&(t?(n._stashedText=n.nodeValue,n.nodeValue=""):n.nodeValue=n._stashedText||""),l&&l.nodeType===8)if(n=l.data,n==="/$"){if(e===0)break;e--}else n!=="$"&&n!=="$?"&&n!=="$~"&&n!=="$!"||e++;n=l}while(n)}function Og(e,t,n){if(t=CSS.escape(t)!==t?"r-"+btoa(t).replace(/=/g,""):t,e.style.viewTransitionName=t,n!=null&&(e.style.viewTransitionClass=n),n=getComputedStyle(e),n.display==="inline"){if(t=e.getClientRects(),t.length===1)var l=1;else for(var a=l=0;a<t.length;a++){var i=t[a];0<i.width&&0<i.height&&l++}l===1&&(e=e.style,e.display=t.length===1?"inline-block":"block",e.marginTop="-"+n.paddingTop,e.marginBottom="-"+n.paddingBottom)}}function Dg(e,t){e=e.style,t=t.style;var n=t!=null?t.hasOwnProperty("viewTransitionName")?t.viewTransitionName:t.hasOwnProperty("view-transition-name")?t["view-transition-name"]:null:null;e.viewTransitionName=n==null||typeof n=="boolean"?"":(""+n).trim(),n=t!=null?t.hasOwnProperty("viewTransitionClass")?t.viewTransitionClass:t.hasOwnProperty("view-transition-class")?t["view-transition-class"]:null:null,e.viewTransitionClass=n==null||typeof n=="boolean"?"":(""+n).trim(),e.display==="inline-block"&&(t==null?e.display=e.margin="":(n=t.display,e.display=n==null||typeof n=="boolean"?"":n,n=t.margin,n!=null?e.margin=n:(n=t.hasOwnProperty("marginTop")?t.marginTop:t["margin-top"],e.marginTop=n==null||typeof n=="boolean"?"":n,t=t.hasOwnProperty("marginBottom")?t.marginBottom:t["margin-bottom"],e.marginBottom=t==null||typeof t=="boolean"?"":t)))}function fx(e,t,n){return n=n.ownerDocument.defaultView,{rect:e,abs:t.position==="absolute"||t.position==="fixed",clip:t.clipPath!=="none"||t.overflow!=="visible"||t.filter!=="none"||t.mask!=="none"||t.mask!=="none"||t.borderRadius!=="0px",view:0<=e.bottom&&0<=e.right&&e.top<=n.innerHeight&&e.left<=n.innerWidth}}function Qf(e){var t=e.getBoundingClientRect(),n=getComputedStyle(e);return fx(t,n,e)}function _x(e){return e.documentElement.clientHeight}function mx(e){this.addEventListener("load",e),this.addEventListener("error",e)}function hx(e,t,n,l,a,i,_,y,M){var L=t.nodeType===9?t:t.ownerDocument;try{var X=L.startViewTransition({update:function(){var A=L.defaultView,q=A.navigation&&A.navigation.transition,ve=L.fonts.status;l();var Le=[];if(ve==="loaded"&&(_x(L),L.fonts.status==="loading"&&Le.push(L.fonts.ready)),ve=Le.length,e!==null)for(var mt=e.suspenseyImages,D=0,R=0;R<mt.length;R++){var U=mt[R];if(!U.complete){var Z=U.getBoundingClientRect();if(0<Z.bottom&&0<Z.right&&Z.top<A.innerHeight&&Z.left<A.innerWidth){if(D+=n0(U),D>Cu){Le.length=ve;break}U=new Promise(mx.bind(U)),Le.push(U)}}}if(0<Le.length)return A=Promise.race([Promise.all(Le),new Promise(function(ze){return setTimeout(ze,500)})]).then(a,a),(q?Promise.allSettled([q.finished,A]):A).then(i,i);if(a(),q)return q.finished.then(i,i);i()},types:n});L.__reactViewTransition=X;var J=[];return X.ready.then(function(){for(var A=L.documentElement.getAnimations({subtree:!0}),q=0;q<A.length;q++){var ve=A[q],Le=ve.effect,mt=Le.pseudoElement;if(mt!=null&&mt.startsWith("::view-transition")){J.push(ve),ve=Le.getKeyframes();for(var D=mt=void 0,R=!0,U=0;U<ve.length;U++){var Z=ve[U],ze=Z.width;if(mt===void 0)mt=ze;else if(mt!==ze){R=!1;break}if(ze=Z.height,D===void 0)D=ze;else if(D!==ze){R=!1;break}delete Z.width,delete Z.height,Z.transform==="none"&&delete Z.transform}R&&mt!==void 0&&D!==void 0&&(Le.setKeyframes(ve),R=getComputedStyle(Le.target,Le.pseudoElement),R.width!==mt||R.height!==D)&&(R=ve[0],R.width=mt,R.height=D,R=ve[ve.length-1],R.width=mt,R.height=D,Le.setKeyframes(ve))}}_()},function(A){L.__reactViewTransition===X&&(L.__reactViewTransition=null);try{if(typeof A=="object"&&A!==null)switch(A.name){case"InvalidStateError":(A.message==="View transition was skipped because document visibility state is hidden."||A.message==="Skipping view transition because document visibility state has become hidden."||A.message==="Skipping view transition because viewport size changed."||A.message==="Transition was aborted because of invalid state")&&(A=null)}A!==null&&M(A)}finally{l(),a(),_()}}),X.finished.finally(function(){for(var A=0;A<J.length;A++)J[A].cancel();L.__reactViewTransition===X&&(L.__reactViewTransition=null),y()}),X}catch{return l(),a(),_(),null}}function Oa(e,t){this._scope=document.documentElement,this._selector="::view-transition-"+e+"("+t+")"}Oa.prototype.animate=function(e,t){return t=typeof t=="number"?{duration:t}:W({},t),t.pseudoElement=this._selector,this._scope.animate(e,t)},Oa.prototype.getAnimations=function(){for(var e=this._scope,t=this._selector,n=e.getAnimations({subtree:!0}),l=[],a=0;a<n.length;a++){var i=n[a].effect;i!==null&&i.target===e&&i.pseudoElement===t&&l.push(n[a])}return l},Oa.prototype.getComputedStyle=function(){return getComputedStyle(this._scope,this._selector)};function Lg(e){return{name:e,group:new Oa("group",e),imagePair:new Oa("image-pair",e),old:new Oa("old",e),new:new Oa("new",e)}}function Il(e){this._fragmentFiber=e,this._observers=this._eventListeners=null}Il.prototype.addEventListener=function(e,t,n){var l=null,a=null;if(!(n!=null&&typeof n!="boolean"&&(l=n.signal||null,l!==null&&l.aborted))){this._eventListeners===null&&(this._eventListeners=[]);var i=this._eventListeners;if($g(i,e,t,n)===-1){var _=this,y=t;n!=null&&typeof n!="boolean"&&n.once===!0&&(y=function(M){_.removeEventListener(e,t,n),typeof t=="function"?t.call(this,M):t.handleEvent(M)}),l!==null&&(a=_.removeEventListener.bind(_,e,t,n),l.addEventListener("abort",a,{once:!0}),a=l.removeEventListener.bind(l,"abort",a)),l=Rr(n),i.push({type:e,listener:t,optionsOrUseCapture:n,attachedListener:y,cleanup:a}),b(this._fragmentFiber.child,!1,gx,e,y,l)}this._eventListeners=i}};function gx(e,t,n,l){return B(e).addEventListener(t,n,l),!1}Il.prototype.removeEventListener=function(e,t,n){var l=this._eventListeners;if(l!==null&&(t=$g(l,e,t,n),t!==-1)){var a=l[t];n=a.attachedListener;var i=a.cleanup;a=Rr(a.optionsOrUseCapture),b(this._fragmentFiber.child,!1,px,e,n,a),l.splice(t,1),i!==null&&i()}};function px(e,t,n,l){return B(e).removeEventListener(t,n,l),!1}function Rr(e){return e!=null&&typeof e!="boolean"&&(e.once===!0||e.signal instanceof AbortSignal)?{capture:e.capture,passive:e.passive}:e}function Bg(e){return e==null?"c=0":typeof e=="boolean"?"c="+(e?"1":"0"):"c="+(e.capture?"1":"0")}function $g(e,t,n,l){if(e.length===0)return-1;l=Bg(l);for(var a=0;a<e.length;a++){var i=e[a];if(i.type===t&&i.listener===n&&Bg(i.optionsOrUseCapture)===l)return a}return-1}Il.prototype.dispatchEvent=function(e){var t=v(this._fragmentFiber);if(t===null)return!0;t=B(t);var n=this._eventListeners;if(n!==null&&0<n.length||!e.bubbles){var l=t.nodeType===9?t.createComment(""):document.createTextNode("");if(n)for(var a=0;a<n.length;a++){var i=n[a];l.addEventListener(i.type,i.attachedListener,Rr(i.optionsOrUseCapture))}if(t.appendChild(l),e=l.dispatchEvent(e),n)for(a=0;a<n.length;a++)i=n[a],l.removeEventListener(i.type,i.attachedListener,Rr(i.optionsOrUseCapture));return t.removeChild(l),e}return t.dispatchEvent(e)},Il.prototype.focus=function(e){b(this._fragmentFiber.child,!0,Ug,e,void 0,void 0)};function Ug(e,t){return e.tag===6?!1:(e=B(e),Ex(e,t))}Il.prototype.focusLast=function(e){var t=[];b(this._fragmentFiber.child,!0,Gf,t,void 0,void 0);for(var n=t.length-1;0<=n&&!Ug(t[n],e);n--);};function Gf(e,t){return t.push(e),!1}Il.prototype.blur=function(){var e=v(this._fragmentFiber);e!==null&&(e=B(e),e=Xi(e).activeElement,e!==null&&b(this._fragmentFiber.child,!1,yx,e,void 0,void 0))};function yx(e,t){return e.tag===6?!1:(e=B(e),e===t||e.contains(t)?(t.blur(),!0):!1)}Il.prototype.observeUsing=function(e){this._observers===null&&(this._observers=new Set),this._observers.add(e),b(this._fragmentFiber.child,!1,xx,e,void 0,void 0)};function xx(e,t){return e.tag===6||(e=B(e),t.observe(e)),!1}Il.prototype.unobserveUsing=function(e){var t=this._observers;if(t!==null&&t.has(e)){t.delete(e),b(this._fragmentFiber.child,!1,bx,e,void 0,void 0);for(var n=t=0;n<_o.length;n++){var l=_o[n];l.fragmentInstance===this&&l.observer===e?e.unobserve(l.instance):_o[t++]=l}_o.length=t}};function bx(e,t){return e.tag===6||(e=B(e),t.unobserve(e)),!1}var _o=[],Vf=!1;function vx(e,t,n){_o.push({fragmentInstance:e,observer:t,instance:n}),Vf||(Vf=!0,Tx(function(){Vf=!1;var l=_o;_o=[];for(var a=0;a<l.length;a++){var i=l[a];i.observer.unobserve(i.instance)}}))}Il.prototype.getClientRects=function(){var e=[];return b(this._fragmentFiber.child,!1,wx,e,void 0,void 0),e};function wx(e,t){if(e.tag===6){e=e.stateNode;var n=e.ownerDocument.createRange();n.selectNodeContents(e),t.push.apply(t,n.getClientRects())}else e=B(e),t.push.apply(t,e.getClientRects());return!1}Il.prototype.getRootNode=function(e){var t=v(this._fragmentFiber);return t===null?this:B(t).getRootNode(e)},Il.prototype.compareDocumentPosition=function(e){var t=v(this._fragmentFiber);if(t===null)return Node.DOCUMENT_POSITION_DISCONNECTED;var n=[];b(this._fragmentFiber.child,!1,Gf,n,void 0,void 0);var l=B(t);if(n.length===0){if(n=l,z(this._fragmentFiber)){e:{for(t=this._fragmentFiber.return;t!==null;){if(t.tag===4){t=t.stateNode.containerInfo;break e}if(t.tag===3||t.tag===5||t.tag===27)break;t=t.return}t=null}t!=null&&(n=t)}t=this._fragmentFiber;var a=l=n.compareDocumentPosition(e);return n===e?a=Node.DOCUMENT_POSITION_CONTAINS:l&Node.DOCUMENT_POSITION_CONTAINED_BY&&(n=E(t)[1],n===null?a=Node.DOCUMENT_POSITION_PRECEDING:(e=B(n).compareDocumentPosition(e),a=e===0||e&Node.DOCUMENT_POSITION_FOLLOWING?Node.DOCUMENT_POSITION_FOLLOWING:Node.DOCUMENT_POSITION_PRECEDING)),a|=Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC}t=B(n[0]),a=B(n[n.length-1]);var i=z(this._fragmentFiber)?t.parentElement:l;if(i==null)return Node.DOCUMENT_POSITION_DISCONNECTED;l=i.compareDocumentPosition(t)&Node.DOCUMENT_POSITION_CONTAINED_BY,i=i.compareDocumentPosition(a)&Node.DOCUMENT_POSITION_CONTAINED_BY;var _=t.compareDocumentPosition(e),y=a.compareDocumentPosition(e),M=_&Node.DOCUMENT_POSITION_CONTAINED_BY||y&Node.DOCUMENT_POSITION_CONTAINED_BY;return y=l&&i&&_&Node.DOCUMENT_POSITION_FOLLOWING&&y&Node.DOCUMENT_POSITION_PRECEDING,t=l&&t===e||i&&a===e||M||y?Node.DOCUMENT_POSITION_CONTAINED_BY:!l&&t===e||!i&&a===e?Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC:_,t&Node.DOCUMENT_POSITION_DISCONNECTED||t&Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC||Sx(t,this._fragmentFiber,n[0],n[n.length-1],e)?t:Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC};function Sx(e,t,n,l,a){var i=yo(a);if(e&Node.DOCUMENT_POSITION_CONTAINED_BY){if(n=!!i)e:{for(;i!==null;){if(i.tag===7&&(i===t||i.alternate===t)){n=!0;break e}i=i.return}n=!1}return n}if(e&Node.DOCUMENT_POSITION_CONTAINS){if(i===null)return i=a.ownerDocument,a===i||a===i.documentElement||a===i.body;e:{for(i=t,t=v(t);i!==null;){if(!(i.tag!==5&&i.tag!==3&&i.tag!==27||i!==t&&i.alternate!==t)){i=!0;break e}i=i.return}i=!1}return i}return e&Node.DOCUMENT_POSITION_PRECEDING?((t=!!i)&&!(t=i===n)&&(t=de(n,i,ie),t===null?t=!1:(b(t,!0,te,i,n),i=F,F=null,t=i!==null)),t):e&Node.DOCUMENT_POSITION_FOLLOWING?((t=!!i)&&!(t=i===l)&&(t=de(l,i,ie),t===null?t=!1:(b(t,!0,C,i,l),i=F,Me=F=null,t=i!==null)),t):!1}function Hg(e,t){var n=e.ownerDocument.createRange();n.selectNodeContents(e),e=n.getBoundingClientRect(),window.scrollTo(window.scrollX+e.left,t?window.scrollY+e.top:window.scrollY+e.bottom-window.innerHeight)}Il.prototype.scrollIntoView=function(e){if(typeof e=="object")throw Error(u(566));var t=[];b(this._fragmentFiber.child,!1,Gf,t,void 0,void 0);var n=e!==!1;if(t.length===0){var l=E(this._fragmentFiber);if(l=n?l[1]||l[0]||v(this._fragmentFiber):l[0]||l[1],l===null)return;if(l.tag===6){e=B(l),Hg(e,n);return}if(l=B(l),l.nodeType!==9){if(l.nodeType===11){n="host"in l?l.host:null,n!==null&&n.scrollIntoView(e);return}l.scrollIntoView(e)}}for(l=n?t.length-1:0;l!==(n?-1:t.length);){var a=t[l];a.tag===6?(a=B(a),Hg(a,n)):B(a).scrollIntoView(e),l+=n?-1:1}};function jx(e,t){return e=B(e),Yg(e,t),!1}function Yg(e,t){e.reactFragments==null&&(e.reactFragments=new Set),e.reactFragments.add(t)}function qg(e,t){var n=t._eventListeners;if(n!==null)for(var l=0;l<n.length;l++){var a=n[l];e.addEventListener(a.type,a.attachedListener,Rr(a.optionsOrUseCapture))}e.nodeType!==3&&(n=t._observers,n!==null&&n.forEach(function(i){for(var _=0,y=0;y<_o.length;y++){var M=_o[y];(M.fragmentInstance!==t||M.observer!==i||M.instance!==e)&&(_o[_++]=M)}_o.length=_,i.observe(e)}),Yg(e,t))}function kx(e,t){var n=t._eventListeners;if(n!==null)for(var l=0;l<n.length;l++){var a=n[l];e.removeEventListener(a.type,a.attachedListener,Rr(a.optionsOrUseCapture))}e.nodeType!==3&&(n=t._observers,n!==null&&n.forEach(function(i){typeof i.rootMargin=="string"?vx(t,i,e):i.unobserve(e)}),e.reactFragments!=null&&e.reactFragments.delete(t))}function Ff(e){var t=e.firstChild;for(t&&t.nodeType===10&&(t=t.nextSibling);t;){var n=t;switch(t=t.nextSibling,n.nodeName){case"HTML":case"HEAD":case"BODY":Ff(n),ea(n);continue;case"SCRIPT":case"STYLE":continue;case"LINK":if(n.rel.toLowerCase()==="stylesheet")continue}e.removeChild(n)}}function Cx(e,t,n,l){for(;e.nodeType===1;){var a=n;if(e.nodeName.toLowerCase()!==t.toLowerCase()){if(!l&&(e.nodeName!=="INPUT"||e.type!=="hidden"))break}else if(l){if(!e[Uo])switch(t){case"meta":if(!e.hasAttribute("itemprop"))break;return e;case"link":if(i=e.getAttribute("rel"),i==="stylesheet"&&e.hasAttribute("data-precedence"))break;if(i!==a.rel||e.getAttribute("href")!==(a.href==null||a.href===""?null:a.href)||e.getAttribute("crossorigin")!==(a.crossOrigin==null?null:a.crossOrigin)||e.getAttribute("title")!==(a.title==null?null:a.title))break;return e;case"style":if(e.hasAttribute("data-precedence"))break;return e;case"script":if(i=e.getAttribute("src"),(i!==(a.src==null?null:a.src)||e.getAttribute("type")!==(a.type==null?null:a.type)||e.getAttribute("crossorigin")!==(a.crossOrigin==null?null:a.crossOrigin))&&i&&e.hasAttribute("async")&&!e.hasAttribute("itemprop"))break;return e;default:return e}}else if(t==="input"&&e.type==="hidden"){var i=a.name==null?null:""+a.name;if(a.type==="hidden"&&e.getAttribute("name")===i)return e}else return e;if(e=Jl(e.nextSibling),e===null)break}return null}function Nx(e,t,n){if(t==="")return null;for(;e.nodeType!==3;)if((e.nodeType!==1||e.nodeName!=="INPUT"||e.type!=="hidden")&&!n||(e=Jl(e.nextSibling),e===null))return null;return e}function Ig(e,t){for(;e.nodeType!==8;)if((e.nodeType!==1||e.nodeName!=="INPUT"||e.type!=="hidden")&&!t||(e=Jl(e.nextSibling),e===null))return null;return e}function Wf(e){return e.data==="$?"||e.data==="$~"}function Kf(e){return e.data==="$!"||e.data==="$?"&&e.ownerDocument.readyState!=="loading"}function Mx(e,t){var n=e.ownerDocument;if(e.data==="$~")e._reactRetry=t;else if(e.data!=="$?"||n.readyState!=="loading")t();else{var l=function(){t(),n.removeEventListener("DOMContentLoaded",l)};n.addEventListener("DOMContentLoaded",l),e._reactRetry=l}}function Jl(e){for(;e!=null;e=e.nextSibling){var t=e.nodeType;if(t===1||t===3)break;if(t===8){if(t=e.data,t==="$"||t==="$!"||t==="$?"||t==="$~"||t==="&"||t==="F!"||t==="F")break;if(t==="/$"||t==="/&")return null}}return e}var Zf=null;function Xg(e){e=e.nextSibling;for(var t=0;e;){if(e.nodeType===8){var n=e.data;if(n==="/$"||n==="/&"){if(t===0)return Jl(e.nextSibling);t--}else n!=="$"&&n!=="$!"&&n!=="$?"&&n!=="$~"&&n!=="&"||t++}e=e.nextSibling}return null}function Qg(e){e=e.previousSibling;for(var t=0;e;){if(e.nodeType===8){var n=e.data;if(n==="$"||n==="$!"||n==="$?"||n==="$~"||n==="&"){if(t===0)return e;t--}else n!=="/$"&&n!=="/&"||t++}e=e.previousSibling}return null}function Ex(e,t){function n(){l=!0}if(e.ownerDocument.activeElement===e)return!0;var l=!1;try{e.ownerDocument.addEventListener("focus",n,!0),(e.focus||HTMLElement.prototype.focus).call(e,t)}finally{e.ownerDocument.removeEventListener("focus",n,!0)}return l}function Tx(e){Rg(function(){Rg(function(t){return e(t)})})}function Gg(e,t,n){switch(t=Xi(n),e){case"html":if(e=t.documentElement,!e)throw Error(u(452));return e;case"head":if(e=t.head,!e)throw Error(u(453));return e;case"body":if(e=t.body,!e)throw Error(u(454));return e;default:throw Error(u(451))}}function Vg(e,t,n){for(var l in n){var a=n[l];n.hasOwnProperty(l)&&a!=null&&Jt(e,t,l,null,sx,a)}n.dangerouslySetInnerHTML!=null&&(e.textContent=""),e.onclick===no&&(e.onclick=null),ea(e)}function Jf(e){for(var t=e.attributes;t.length;)e.removeAttributeNode(t[0]);ea(e)}var Pl=new Map,Fg=new Set;function Qi(e){if(typeof e.getRootNode=="function"){var t=e.getRootNode();if(t.nodeType===9||t.nodeType===11)return t}return e.nodeType===9?e:e.ownerDocument}var as=Ue.d;Ue.d={f:Rx,r:zx,D:Ax,C:Ox,L:Dx,m:Lx,X:$x,S:Bx,M:Ux};function Rx(){var e=as.f(),t=pu();return e||t}function zx(e){var t=xo(e);t!==null&&t.tag===5&&t.type==="form"?Km(t):as.r(e)}var zr=typeof document>"u"?null:document;function Wg(e,t,n){var l=zr;if(l&&typeof t=="string"&&t){var a=Xn(t);a='link[rel="'+e+'"][href="'+a+'"]',typeof n=="string"&&(a+='[crossorigin="'+n+'"]'),Fg.has(a)||(Fg.add(a),e={rel:e,crossOrigin:n,href:t},l.querySelector(a)===null&&(t=l.createElement("link"),nl(t,"link",e),Tn(t),l.head.appendChild(t)))}}function Ax(e){as.D(e),Wg("dns-prefetch",e,null)}function Ox(e,t){as.C(e,t),Wg("preconnect",e,t)}function Dx(e,t,n){as.L(e,t,n);var l=zr;if(l&&e&&t){var a='link[rel="preload"][as="'+Xn(t)+'"]';t==="image"&&n&&n.imageSrcSet?(a+='[imagesrcset="'+Xn(n.imageSrcSet)+'"]',typeof n.imageSizes=="string"&&(a+='[imagesizes="'+Xn(n.imageSizes)+'"]')):a+='[href="'+Xn(e)+'"]';var i=a;switch(t){case"style":i=Ar(e);break;case"script":i=Or(e)}if(!(Pl.has(i)||(e=W({rel:"preload",href:t==="image"&&n&&n.imageSrcSet?void 0:e,as:t},n),Pl.set(i,e),l.querySelector(a)!==null||t==="style"&&l.querySelector(Gi(i))||t==="script"&&l.querySelector(Vi(i))))){var _=l.createElement("link");nl(_,"link",e),t==="style"&&(_[Yt]=!0,_.onload=_.onerror=function(){ad(_)}),Tn(_),l.head.appendChild(_)}}}function Lx(e,t){as.m(e,t);var n=zr;if(n&&e){var l=t&&typeof t.as=="string"?t.as:"script",a='link[rel="modulepreload"][as="'+Xn(l)+'"][href="'+Xn(e)+'"]',i=a;switch(l){case"audioworklet":case"paintworklet":case"serviceworker":case"sharedworker":case"worker":case"script":i=Or(e)}if(!Pl.has(i)&&(e=W({rel:"modulepreload",href:e},t),Pl.set(i,e),n.querySelector(a)===null)){switch(l){case"audioworklet":case"paintworklet":case"serviceworker":case"sharedworker":case"worker":case"script":if(n.querySelector(Vi(i)))return}l=n.createElement("link"),nl(l,"link",e),Tn(l),n.head.appendChild(l)}}}function Bx(e,t,n){as.S(e,t,n);var l=zr;if(l&&e){var a=us(l).hoistableStyles,i=Ar(e);t=t||"default";var _=a.get(i);if(!_){var y={loading:0,preload:null};if(_=l.querySelector(Gi(i)))y.loading=5;else{e=W({rel:"stylesheet",href:e,"data-precedence":t},n),(n=Pl.get(i))&&Pf(e,n);var M=_=l.createElement("link");Tn(M),nl(M,"link",e),M._p=new Promise(function(L,X){M.onload=L,M.onerror=X}),M.addEventListener("load",function(){y.loading|=1}),M.addEventListener("error",function(){y.loading|=2}),y.loading|=4,ju(_,t,l)}_={type:"stylesheet",instance:_,count:1,state:y},a.set(i,_)}}}function $x(e,t){as.X(e,t);var n=zr;if(n&&e){var l=us(n).hoistableScripts,a=Or(e),i=l.get(a);i||(i=n.querySelector(Vi(a)),i||(e=W({src:e,async:!0},t),(t=Pl.get(a))&&e_(e,t),i=n.createElement("script"),Tn(i),nl(i,"link",e),n.head.appendChild(i)),i={type:"script",instance:i,count:1,state:null},l.set(a,i))}}function Ux(e,t){as.M(e,t);var n=zr;if(n&&e){var l=us(n).hoistableScripts,a=Or(e),i=l.get(a);i||(i=n.querySelector(Vi(a)),i||(e=W({src:e,async:!0,type:"module"},t),(t=Pl.get(a))&&e_(e,t),i=n.createElement("script"),Tn(i),nl(i,"link",e),n.head.appendChild(i)),i={type:"script",instance:i,count:1,state:null},l.set(a,i))}}function Kg(e,t,n,l){var a=(a=lt.current)?Qi(a):null;if(!a)throw Error(u(446));switch(e){case"meta":case"title":return null;case"style":return typeof n.precedence=="string"&&typeof n.href=="string"?(n=Ar(n.href),t=us(a).hoistableStyles,l=t.get(n),l||(l={type:"style",instance:null,count:0,state:null},t.set(n,l)),l):{type:"void",instance:null,count:0,state:null};case"link":if(n.rel==="stylesheet"&&typeof n.href=="string"&&typeof n.precedence=="string"){e=Ar(n.href);var i=us(a).hoistableStyles,_=i.get(e);if(_||(a=a.ownerDocument||a,_={type:"stylesheet",instance:null,count:0,state:{loading:0,preload:null}},i.set(e,_),(i=a.querySelector(Gi(e)))?i._p||(_.instance=i,_.state.loading=5):(i=Pl.get(e),i||(i={rel:"preload",as:"style",href:n.href,crossOrigin:n.crossOrigin,integrity:n.integrity,media:n.media,hrefLang:n.hrefLang,referrerPolicy:n.referrerPolicy},Pl.set(e,i)),Hx(a,e,i,_.state))),t&&l===null)throw Error(u(528,""));return _}if(t&&l!==null)throw Error(u(529,""));return null;case"script":return t=n.async,n=n.src,typeof n=="string"&&t&&typeof t!="function"&&typeof t!="symbol"?(n=Or(n),t=us(a).hoistableScripts,l=t.get(n),l||(l={type:"script",instance:null,count:0,state:null},t.set(n,l)),l):{type:"void",instance:null,count:0,state:null};default:throw Error(u(444,e))}}function Ar(e){return'href="'+Xn(e)+'"'}function Gi(e){return'link[rel="stylesheet"]['+e+"]"}function Zg(e){return W({},e,{"data-precedence":e.precedence,precedence:null})}function Hx(e,t,n,l){if(t=e.querySelector('link[rel="preload"][as="style"]['+t+"]")){if(t[Yt]!==!0){l.loading=1;return}}else t=e.createElement("link"),t[Yt]=!0,t.onload=t.onerror=ad.bind(null,t),nl(t,"link",n),Tn(t),e.head.appendChild(t);l.preload=t,t.addEventListener("load",function(){return l.loading|=1}),t.addEventListener("error",function(){return l.loading|=2})}function Or(e){return'[src="'+Xn(e)+'"]'}function Vi(e){return"script[async]"+e}function Jg(e,t,n){if(t.count++,t.instance===null)switch(t.type){case"style":var l=e.querySelector('style[data-href~="'+Xn(n.href)+'"]');if(l)return t.instance=l,Tn(l),l;var a=W({},n,{"data-href":n.href,"data-precedence":n.precedence,href:null,precedence:null});return l=(e.ownerDocument||e).createElement("style"),Tn(l),nl(l,"style",a),ju(l,n.precedence,e),t.instance=l;case"stylesheet":a=Ar(n.href);var i=e.querySelector(Gi(a));if(i)return t.state.loading|=4,t.instance=i,Tn(i),i;l=Zg(n),(a=Pl.get(a))&&Pf(l,a),i=(e.ownerDocument||e).createElement("link"),Tn(i);var _=i;return _._p=new Promise(function(y,M){_.onload=y,_.onerror=M}),nl(i,"link",l),t.state.loading|=4,ju(i,n.precedence,e),t.instance=i;case"script":return i=Or(n.src),(a=e.querySelector(Vi(i)))?(t.instance=a,Tn(a),a):(l=n,(a=Pl.get(i))&&(l=W({},n),e_(l,a)),e=e.ownerDocument||e,a=e.createElement("script"),Tn(a),nl(a,"link",l),e.head.appendChild(a),t.instance=a);case"void":return null;default:throw Error(u(443,t.type))}else t.type==="stylesheet"&&(t.state.loading&4)===0&&(l=t.instance,t.state.loading|=4,ju(l,n.precedence,e));return t.instance}function ju(e,t,n){for(var l=n.querySelectorAll('link[rel="stylesheet"][data-precedence],style[data-precedence]'),a=l.length?l[l.length-1]:null,i=a,_=0;_<l.length;_++){var y=l[_];if(y.dataset.precedence===t)i=y;else if(i!==a)break}i?i.parentNode.insertBefore(e,i.nextSibling):(t=n.nodeType===9?n.head:n,t.insertBefore(e,t.firstChild))}function Pf(e,t){e.crossOrigin==null&&(e.crossOrigin=t.crossOrigin),e.referrerPolicy==null&&(e.referrerPolicy=t.referrerPolicy),e.title==null&&(e.title=t.title)}function e_(e,t){e.crossOrigin==null&&(e.crossOrigin=t.crossOrigin),e.referrerPolicy==null&&(e.referrerPolicy=t.referrerPolicy),e.integrity==null&&(e.integrity=t.integrity)}var ku=null;function Pg(e,t,n){if(ku===null){var l=new Map,a=ku=new Map;a.set(n,l)}else a=ku,l=a.get(n),l||(l=new Map,a.set(n,l));if(l.has(e))return l;for(l.set(e,null),n=n.getElementsByTagName(e),a=0;a<n.length;a++){var i=n[a];if(!(i[Uo]||i[En]||e==="link"&&i.getAttribute("rel")==="stylesheet")&&i.namespaceURI!=="http://www.w3.org/2000/svg"){var _=i.getAttribute(t)||"";_=e+_;var y=l.get(_);y?y.push(i):l.set(_,[i])}}return l}function t_(e,t,n){e=e.ownerDocument||e,e.head.insertBefore(n,t==="title"?e.querySelector("head > title"):null)}function Yx(e,t,n){if(n===1||t.itemProp!=null)return!1;switch(e){case"meta":case"title":return!0;case"style":if(typeof t.precedence!="string"||typeof t.href!="string"||t.href==="")break;return!0;case"link":if(typeof t.rel!="string"||typeof t.href!="string"||t.href===""||t.onLoad||t.onError)break;switch(t.rel){case"stylesheet":return e=t.disabled,typeof t.precedence=="string"&&e==null;default:return!0}case"script":if(t.async&&typeof t.async!="function"&&typeof t.async!="symbol"&&!t.onLoad&&!t.onError&&t.src&&typeof t.src=="string")return!0}return!1}function e0(e,t){return e==="img"&&t.src!=null&&t.src!==""&&t.onLoad==null&&t.loading!=="lazy"}function t0(e){return!(e.type==="stylesheet"&&(e.state.loading&3)===0)}function n0(e){return(e.width||100)*(e.height||100)*(typeof devicePixelRatio=="number"?devicePixelRatio:1)*.25}function l0(e,t){typeof t.decode=="function"&&(e.imgCount++,t.complete||(e.imgBytes+=n0(t),e.suspenseyImages.push(t)),e=Xx.bind(e),t.decode().then(e,e))}function qx(e,t,n,l){if(n.type==="stylesheet"&&(typeof l.media!="string"||matchMedia(l.media).matches!==!1)&&(n.state.loading&4)===0){if(n.instance===null){var a=Ar(l.href),i=t.querySelector(Gi(a));if(i){t=i._p,t!==null&&typeof t=="object"&&typeof t.then=="function"&&(e.count++,e=Fi.bind(e),t.then(e,e)),n.state.loading|=4,n.instance=i,Tn(i);return}i=t.ownerDocument||t,l=Zg(l),(a=Pl.get(a))&&Pf(l,a),i=i.createElement("link"),Tn(i);var _=i;_._p=new Promise(function(y,M){_.onload=y,_.onerror=M}),nl(i,"link",l),n.instance=i}e.stylesheets===null&&(e.stylesheets=new Map),e.stylesheets.set(n,t),(t=n.state.preload)&&(n.state.loading&3)===0&&(e.count++,n=Fi.bind(e),t.addEventListener("load",n),t.addEventListener("error",n))}}var Cu=0;function Ix(e,t){return e.stylesheets&&e.count===0&&Mu(e,e.stylesheets),0<e.count||0<e.imgCount?function(n){var l=setTimeout(function(){if(e.stylesheets&&Mu(e,e.stylesheets),e.unsuspend){var i=e.unsuspend;e.unsuspend=null,i()}},6e4+t);0<e.imgBytes&&Cu===0&&(Cu=62500*rx());var a=setTimeout(function(){if(e.waitingForImages=!1,e.count===0&&(e.stylesheets&&Mu(e,e.stylesheets),e.unsuspend)){var i=e.unsuspend;e.unsuspend=null,i()}},(e.imgBytes>Cu?50:800)+t);return e.unsuspend=n,function(){e.unsuspend=null,clearTimeout(l),clearTimeout(a)}}:null}function o0(e){if(e.count===0&&(e.imgCount===0||!e.waitingForImages)){if(e.stylesheets)Mu(e,e.stylesheets);else if(e.unsuspend){var t=e.unsuspend;e.unsuspend=null,t()}}}function Fi(){this.count--,o0(this)}function Xx(){this.imgCount--,o0(this)}var Nu=null;function Mu(e,t){e.stylesheets=null,e.unsuspend!==null&&(e.count++,Nu=new Map,t.forEach(Qx,e),Nu=null,Fi.call(e))}function Qx(e,t){if(!(t.state.loading&4)){var n=Nu.get(e);if(n)var l=n.get(null);else{n=new Map,Nu.set(e,n);for(var a=e.querySelectorAll("link[data-precedence],style[data-precedence]"),i=0;i<a.length;i++){var _=a[i];(_.nodeName==="LINK"||_.getAttribute("media")!=="not all")&&(n.set(_.dataset.precedence,_),l=_)}l&&n.set(null,l)}a=t.instance,_=a.getAttribute("data-precedence"),i=n.get(_)||l,i===l&&n.set(null,a),n.set(_,a),this.count++,l=Fi.bind(this),a.addEventListener("load",l),a.addEventListener("error",l),i?i.parentNode.insertBefore(a,i.nextSibling):(e=e.nodeType===9?e.head:e,e.insertBefore(a,e.firstChild)),t.state.loading|=4}}var Dr={$$typeof:ue,Provider:null,Consumer:null,_currentValue:un,_currentValue2:un,_threadCount:0};function Gx(e,t,n,l,a,i,_,y,M){this.tag=1,this.containerInfo=e,this.pingCache=this.current=this.pendingChildren=null,this.timeoutHandle=-1,this.callbackNode=this.next=this.pendingContext=this.context=this.cancelPendingCommit=null,this.callbackPriority=0,this.expirationTimes=Ot(-1),this.entangledLanes=this.shellSuspendCounter=this.errorRecoveryDisabledLanes=this.expiredLanes=this.warmLanes=this.pingedLanes=this.suspendedLanes=this.pendingLanes=0,this.entanglements=Ot(0),this.hiddenUpdates=Ot(null),this.identifierPrefix=l,this.onUncaughtError=a,this.onCaughtError=i,this.onRecoverableError=_,this.pooledCache=null,this.pooledCacheLanes=0,this.formState=M,this.transitionTypes=null,this.incompleteTransitions=new Map}function s0(e,t,n,l,a,i,_,y,M,L,X,J){return e=new Gx(e,t,n,_,M,L,X,J,y),t=1,i===!0&&(t|=24),i=He(3,null,null,t),e.current=i,i.stateNode=e,t=hd(),t.refCount++,e.pooledCache=t,t.refCount++,i.memoizedState={element:l,isDehydrated:n,cache:t},xd(i),e}function a0(e){return e?(e=$t,e):$t}function r0(e,t,n,l,a,i){a=a0(a),l.context===null?l.context=a:l.pendingContext=a,l=Rs(t),l.payload={element:n},i=i===void 0?null:i,i!==null&&(l.callback=i),n=zs(e,l,t),n!==null&&(Tl(n,e,t),ki(n,e,t))}function i0(e,t){if(e=e.memoizedState,e!==null&&e.dehydrated!==null){var n=e.retryLane;e.retryLane=n!==0&&n<t?n:t}}function n_(e,t){i0(e,t),(e=e.alternate)&&i0(e,t)}function c0(e){if(e.tag===13||e.tag===31){var t=Ve(e,67108864);t!==null&&Tl(t,e,67108864),n_(e,67108864)}}function u0(e){if(e.tag===13||e.tag===31){var t=ql();t=Wr(t);var n=Ve(e,t);n!==null&&Tl(n,e,t),n_(e,t)}}var Lr=!0;function Vx(e,t,n,l){var a=Re.T;Re.T=null;var i=Ue.p;try{Ue.p=2,l_(e,t,n,l)}finally{Ue.p=i,Re.T=a}}function Fx(e,t,n,l){var a=Re.T;Re.T=null;var i=Ue.p;try{Ue.p=8,l_(e,t,n,l)}finally{Ue.p=i,Re.T=a}}function l_(e,t,n,l){if(Lr){var a=o_(l);if(a===null)$f(e,t,l,Eu,n),f0(e,l);else if(Kx(a,e,t,n,l))l.stopPropagation();else if(f0(e,l),t&4&&-1<Wx.indexOf(e)){for(;a!==null;){var i=xo(a);if(i!==null)switch(i.tag){case 3:if(i=i.stateNode,i.current.memoizedState.isDehydrated){var _=Mn(i.pendingLanes);if(_!==0){var y=i;for(y.pendingLanes|=2,y.entangledLanes|=2;_;){var M=1<<31-Gt(_);y.entanglements[1]|=M,_&=~M}Bo(i),(qt&6)===0&&(mu=Ze()+500,Yi(0))}}break;case 31:case 13:y=Ve(i,2),y!==null&&Tl(y,i,2),pu(),n_(i,2)}if(i=o_(l),i===null&&$f(e,t,l,Eu,n),i===a)break;a=i}a!==null&&l.stopPropagation()}else $f(e,t,l,null,n)}}function o_(e){return e=sa(e),s_(e)}var Eu=null;function s_(e){if(Eu=null,e=yo(e),e!==null){var t=d(e);if(t===null)e=null;else{var n=t.tag;if(n===13){if(e=m(t),e!==null)return e;e=null}else if(n===31){if(e=g(t),e!==null)return e;e=null}else if(n===3){if(t.stateNode.current.memoizedState.isDehydrated)return t.tag===3?t.stateNode.containerInfo:null;e=null}else t!==e&&(e=null)}}return Eu=e,null}function d0(e){switch(e){case"beforetoggle":case"cancel":case"click":case"close":case"contextmenu":case"copy":case"cut":case"auxclick":case"dblclick":case"dragend":case"dragstart":case"drop":case"focusin":case"focusout":case"input":case"invalid":case"keydown":case"keypress":case"keyup":case"mousedown":case"mouseup":case"paste":case"pause":case"play":case"pointercancel":case"pointerdown":case"pointerup":case"ratechange":case"reset":case"seeked":case"submit":case"toggle":case"touchcancel":case"touchend":case"touchstart":case"volumechange":case"change":case"selectionchange":case"textInput":case"compositionstart":case"compositionend":case"compositionupdate":case"beforeblur":case"afterblur":case"beforeinput":case"blur":case"fullscreenchange":case"fullscreenerror":case"focus":case"hashchange":case"popstate":case"select":case"selectstart":return 2;case"drag":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"mousemove":case"mouseout":case"mouseover":case"pointermove":case"pointerout":case"pointerover":case"resize":case"scroll":case"touchmove":case"wheel":case"mouseenter":case"mouseleave":case"pointerenter":case"pointerleave":return 8;case"message":switch(bt()){case Bt:return 2;case pn:return 8;case pt:case At:return 32;case kt:return 268435456;default:return 32}default:return 32}}var a_=!1,Xs=null,Qs=null,Gs=null,Wi=new Map,Ki=new Map,Vs=[],Wx="mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(" ");function f0(e,t){switch(e){case"focusin":case"focusout":Xs=null;break;case"dragenter":case"dragleave":Qs=null;break;case"mouseover":case"mouseout":Gs=null;break;case"pointerover":case"pointerout":Wi.delete(t.pointerId);break;case"gotpointercapture":case"lostpointercapture":Ki.delete(t.pointerId)}}function Zi(e,t,n,l,a,i){return e===null||e.nativeEvent!==i?(e={blockedOn:t,domEventName:n,eventSystemFlags:l,nativeEvent:i,targetContainers:[a]},t!==null&&(t=xo(t),t!==null&&c0(t)),e):(e.eventSystemFlags|=l,t=e.targetContainers,a!==null&&t.indexOf(a)===-1&&t.push(a),e)}function Kx(e,t,n,l,a){switch(t){case"focusin":return Xs=Zi(Xs,e,t,n,l,a),!0;case"dragenter":return Qs=Zi(Qs,e,t,n,l,a),!0;case"mouseover":return Gs=Zi(Gs,e,t,n,l,a),!0;case"pointerover":var i=a.pointerId;return Wi.set(i,Zi(Wi.get(i)||null,e,t,n,l,a)),!0;case"gotpointercapture":return i=a.pointerId,Ki.set(i,Zi(Ki.get(i)||null,e,t,n,l,a)),!0}return!1}function _0(e){var t=yo(e.target);if(t!==null){var n=d(t);if(n!==null){if(t=n.tag,t===13){if(t=m(n),t!==null){e.blockedOn=t,Ua(e.priority,function(){u0(n)});return}}else if(t===31){if(t=g(n),t!==null){e.blockedOn=t,Ua(e.priority,function(){u0(n)});return}}else if(t===3&&n.stateNode.current.memoizedState.isDehydrated){e.blockedOn=n.tag===3?n.stateNode.containerInfo:null;return}}}e.blockedOn=null}function Tu(e){if(e.blockedOn!==null)return!1;for(var t=e.targetContainers;0<t.length;){var n=o_(e.nativeEvent);if(n===null){n=e.nativeEvent;var l=new n.constructor(n.type,n);oa=l,n.target.dispatchEvent(l),oa=null}else return t=xo(n),t!==null&&c0(t),e.blockedOn=n,!1;t.shift()}return!0}function m0(e,t,n){Tu(e)&&n.delete(t)}function Zx(){a_=!1,Xs!==null&&Tu(Xs)&&(Xs=null),Qs!==null&&Tu(Qs)&&(Qs=null),Gs!==null&&Tu(Gs)&&(Gs=null),Wi.forEach(m0),Ki.forEach(m0)}function Ru(e,t){e.blockedOn===t&&(e.blockedOn=null,a_||(a_=!0,s.unstable_scheduleCallback(s.unstable_NormalPriority,Zx)))}var zu=null;function h0(e){zu!==e&&(zu=e,s.unstable_scheduleCallback(s.unstable_NormalPriority,function(){zu===e&&(zu=null);for(var t=0;t<e.length;t+=3){var n=e[t],l=e[t+1],a=e[t+2];if(typeof l!="function"){if(s_(l||n)===null)continue;break}var i=xo(n);i!==null&&(e.splice(t,3),t-=3,Hd(i,{pending:!0,data:a,method:n.method,action:l},l,a))}}))}function Br(e){function t(M){return Ru(M,e)}Xs!==null&&Ru(Xs,e),Qs!==null&&Ru(Qs,e),Gs!==null&&Ru(Gs,e),Wi.forEach(t),Ki.forEach(t);for(var n=0;n<Vs.length;n++){var l=Vs[n];l.blockedOn===e&&(l.blockedOn=null)}for(;0<Vs.length&&(n=Vs[0],n.blockedOn===null);)_0(n),n.blockedOn===null&&Vs.shift();if(n=(e.ownerDocument||e).$$reactFormReplay,n!=null)for(l=0;l<n.length;l+=3){var a=n[l],i=n[l+1],_=a[wn]||null;if(typeof i=="function")_||h0(n);else if(_){var y=null;if(i&&i.hasAttribute("formAction")){if(a=i,_=i[wn]||null)y=_.formAction;else if(s_(a)!==null)continue}else y=_.action;typeof y=="function"?n[l+1]=y:(n.splice(l,3),l-=3),h0(n)}}}function g0(){function e(i){i.canIntercept&&i.info==="react-transition"&&i.intercept({handler:function(){return new Promise(function(_){return a=_})},focusReset:"manual",scroll:"manual"})}function t(){a!==null&&(a(),a=null),l||setTimeout(n,20)}function n(){if(!l&&!navigation.transition){var i=navigation.currentEntry;i&&i.url!=null&&navigation.navigate(i.url,{state:i.getState(),info:"react-transition",history:"replace"})}}if(typeof navigation=="object"){var l=!1,a=null;return navigation.addEventListener("navigate",e),navigation.addEventListener("navigatesuccess",t),navigation.addEventListener("navigateerror",t),setTimeout(n,100),function(){l=!0,navigation.removeEventListener("navigate",e),navigation.removeEventListener("navigatesuccess",t),navigation.removeEventListener("navigateerror",t),a!==null&&(a(),a=null)}}}function r_(e){this._internalRoot=e}Au.prototype.render=r_.prototype.render=function(e){var t=this._internalRoot;if(t===null)throw Error(u(409));var n=t.current,l=ql();r0(n,l,e,t,null,null)},Au.prototype.unmount=r_.prototype.unmount=function(){var e=this._internalRoot;if(e!==null){this._internalRoot=null;var t=e.containerInfo;r0(e.current,2,null,e,null,null),pu(),t[Gl]=null}};function Au(e){this._internalRoot=e}Au.prototype.unstable_scheduleHydration=function(e){if(e){var t=Kr();e={blockedOn:null,target:e,priority:t};for(var n=0;n<Vs.length&&t!==0&&t<Vs[n].priority;n++);Vs.splice(n,0,e),n===0&&_0(e)}};var p0=r.version;if(p0!=="19.3.0")throw Error(u(527,p0,"19.3.0"));Ue.findDOMNode=function(e){var t=e._reactInternals;if(t===void 0)throw typeof e.render=="function"?Error(u(188)):(e=Object.keys(e).join(","),Error(u(268,e)));return e=j(t),e=e!==null?k(e):null,e=e===null?null:e.stateNode,e};var Jx={bundleType:0,version:"19.3.0",rendererPackageName:"react-dom",currentDispatcherRef:Re,reconcilerVersion:"19.3.0"};if(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__<"u"){var Ou=__REACT_DEVTOOLS_GLOBAL_HOOK__;if(!Ou.isDisabled&&Ou.supportsFiber)try{vt=Ou.inject(Jx),Je=Ou}catch{}}return Pi.createRoot=function(e,t){if(!f(e))throw Error(u(299));var n=!1,l="",a=ah,i=rh,_=ih;return t!=null&&(t.unstable_strictMode===!0&&(n=!0),t.identifierPrefix!==void 0&&(l=t.identifierPrefix),t.onUncaughtError!==void 0&&(a=t.onUncaughtError),t.onCaughtError!==void 0&&(i=t.onCaughtError),t.onRecoverableError!==void 0&&(_=t.onRecoverableError)),t=s0(e,1,!1,null,null,n,l,null,a,i,_,g0),e[Gl]=t.current,Bf(e),new r_(t)},Pi.hydrateRoot=function(e,t,n){if(!f(e))throw Error(u(299));var l=!1,a="",i=ah,_=rh,y=ih,M=null;return n!=null&&(n.unstable_strictMode===!0&&(l=!0),n.identifierPrefix!==void 0&&(a=n.identifierPrefix),n.onUncaughtError!==void 0&&(i=n.onUncaughtError),n.onCaughtError!==void 0&&(_=n.onCaughtError),n.onRecoverableError!==void 0&&(y=n.onRecoverableError),n.formState!==void 0&&(M=n.formState)),t=s0(e,1,!0,t,n??null,l,a,M,i,_,y,g0),t.context=a0(null),n=t.current,l=ql(),l=Wr(l),a=Rs(l),a.callback=null,zs(n,a,l),n=l,t.current.lanes=n,Ps(t,n),Bo(t),e[Gl]=t.current,Bf(e),new Au(t)},Pi.version="19.3.0",Pi}var j0;function Eb(){if(j0)return u_.exports;j0=1;function s(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__>"u"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(s)}catch(r){console.error(r)}}return s(),u_.exports=Mb(),u_.exports}var Tb=Eb();const Rb=cb(Tb),zb="modulepreload",Ab=function(s){return"/"+s},k0={},_c=function(r,c,u){let f=Promise.resolve();if(c&&c.length>0){let m=function(j){return Promise.all(j.map(k=>Promise.resolve(k).then(b=>({status:"fulfilled",value:b}),b=>({status:"rejected",reason:b}))))};document.getElementsByTagName("link");const g=document.querySelector("meta[property=csp-nonce]"),p=(g==null?void 0:g.nonce)||(g==null?void 0:g.getAttribute("nonce"));f=m(c.map(j=>{if(j=Ab(j),j in k0)return;k0[j]=!0;const k=j.endsWith(".css"),b=k?'[rel="stylesheet"]':"";if(document.querySelector(`link[href="${j}"]${b}`))return;const v=document.createElement("link");if(v.rel=k?"stylesheet":zb,k||(v.as="script"),v.crossOrigin="",v.href=j,p&&v.setAttribute("nonce",p),document.head.appendChild(v),k)return new Promise((z,E)=>{v.addEventListener("load",z),v.addEventListener("error",()=>E(new Error(`Unable to preload CSS for ${j}`)))})}))}function d(m){const g=new Event("vite:preloadError",{cancelable:!0});if(g.payload=m,window.dispatchEvent(g),!g.defaultPrevented)throw m}return f.then(m=>{for(const g of m||[])g.status==="rejected"&&d(g.reason);return r().catch(d)})},Dt="/api";async function Ft(s){if(!s.ok){let r=`HTTP ${s.status}: ${s.statusText}`;try{const c=await s.json();c!=null&&c.error?r=c.error:c!=null&&c.message&&(r=c.message)}catch{}throw new Error(r)}return s.json()}async function Ob(){const s=await fetch(`${Dt}/status`);return Ft(s)}async function Db(s,r){let c,u,f=!1;typeof s=="object"&&s!==null?(c=s.agent,u=s.profile,f=!!s.refresh):(c=s,f=!1);const d=new URLSearchParams;c&&c!=="all"&&d.set("agent",c),u&&u!=="all"&&d.set("profile",u),f&&d.set("refresh","true");const m=d.toString(),g=m?`${Dt}/profiles?${m}`:`${Dt}/profiles`,p=await fetch(g);return Ft(p)}async function Lb(s){const r=await fetch(`${Dt}/profiles`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(s)});return Ft(r)}async function jw(s,r){const c=!!s&&s.trim()!==""&&s!=="all",u=c?`${Dt}/profiles/${encodeURIComponent(s)}/${encodeURIComponent(r)}`:`${Dt}/profiles/${encodeURIComponent(r)}`,f=await fetch(u,{method:"DELETE"});if(!f.ok&&f.status===404){const d=c?`agent=${encodeURIComponent(s)}&name=${encodeURIComponent(r)}`:`name=${encodeURIComponent(r)}`,m=await fetch(`${Dt}/profiles?${d}`,{method:"DELETE"});return m.ok?void 0:Ft(m)}return Ft(f)}async function __(s){const r=new URLSearchParams;s!=null&&s.agent&&r.set("agent",s.agent),s!=null&&s.profile&&r.set("profile",s.profile),s!=null&&s.cwd&&r.set("cwd",s.cwd),s!=null&&s.query&&(r.set("q",s.query),r.set("query",s.query)),s!=null&&s.limit&&r.set("limit",s.limit.toString()),s!=null&&s.force&&r.set("force","true");const c=r.toString(),u=c?`${Dt}/sessions?${c}`:`${Dt}/sessions`,f=await fetch(u);return Ft(f)}async function kw(s){const r=await fetch(`${Dt}/sessions/resume`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(s)});return Ft(r)}async function Cw(s){const r=s?`${Dt}/mcp?profile=${encodeURIComponent(s)}`:`${Dt}/mcp`,c=await fetch(r);return Ft(c)}async function Nw(s,r){const c=await fetch(`${Dt}/profiles/${encodeURIComponent(s)}/config`,{method:"PATCH",headers:{"Content-Type":"application/json"},body:JSON.stringify(r)});return Ft(c)}async function Bb(){const s=await fetch(`${Dt}/daemon`);return Ft(s)}async function C0(s=200,r=0){const c=await fetch(`${Dt}/daemon/logs?tail=${s}&offset=${r}`);return Ft(c)}async function $b(s){const r=await fetch(`${Dt}/daemon/install`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({binary_path:s})});return Ft(r)}async function Ub(){const s=await fetch(`${Dt}/daemon/uninstall`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({})});return Ft(s)}async function Hb(){const s=await fetch(`${Dt}/daemon/run`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({})});return Ft(s)}async function Mw(){const s=await fetch(`${Dt}/adapters`);return Ft(s)}async function Ew(s,r){const c=await fetch(`${Dt}/sessions/${encodeURIComponent(s)}/${encodeURIComponent(r)}/transcript`);return Ft(c)}async function Tw(s=!1){const r=s?`${Dt}/share/status?refresh=true`:`${Dt}/share/status`,c=await fetch(r);return Ft(c)}async function Rw(s){const r=await fetch(`${Dt}/share/approve`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({request_id:s})});return Ft(r)}async function zw(s){const r=await fetch(`${Dt}/share/deny`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({request_id:s})});return Ft(r)}async function Aw(s){const r=await fetch(`${Dt}/share/cancel`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({request_id:s})});return Ft(r)}async function Ow(s){const r=await fetch(`${Dt}/share/poll`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({request_id:s})});return Ft(r)}async function Dw(s){const r=await fetch(`${Dt}/share/revoke`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({lease_id:s})});return Ft(r)}async function Lw(s){const r=await fetch(`${Dt}/share/global`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({mode:s})});return Ft(r)}async function Bw(s){const r=await fetch(`${Dt}/share/toggle`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(s)});return Ft(r)}async function $w(s){const r=await fetch(`${Dt}/share/peers`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({host:s})});return Ft(r)}async function Uw(s){const r=await fetch(`${Dt}/share/request`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(s)});return Ft(r)}async function Hw(s){const r=await fetch(`${Dt}/share/release`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({profile:s})});return Ft(r)}async function Yw(){const s=await fetch(`${Dt}/share/clean-expired`,{method:"POST"});return Ft(s)}async function qw(s){const r=await fetch(`${Dt}/share/schedules`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(s)});return Ft(r)}async function Iw(s){const r=await fetch(`${Dt}/share/schedules`,{method:"DELETE",headers:{"Content-Type":"application/json"},body:JSON.stringify({id:s})});return Ft(r)}function Mp(s){var r,c,u="";if(typeof s=="string"||typeof s=="number")u+=s;else if(typeof s=="object")if(Array.isArray(s)){var f=s.length;for(r=0;r<f;r++)s[r]&&(c=Mp(s[r]))&&(u&&(u+=" "),u+=c)}else for(c in s)s[c]&&(u&&(u+=" "),u+=c);return u}function Ep(){for(var s,r,c=0,u="",f=arguments.length;c<f;c++)(s=arguments[c])&&(r=Mp(s))&&(u&&(u+=" "),u+=r);return u}const B_="-",Yb=s=>{const r=Ib(s),{conflictingClassGroups:c,conflictingClassGroupModifiers:u}=s;return{getClassGroupId:m=>{const g=m.split(B_);return g[0]===""&&g.length!==1&&g.shift(),Tp(g,r)||qb(m)},getConflictingClassGroupIds:(m,g)=>{const p=c[m]||[];return g&&u[m]?[...p,...u[m]]:p}}},Tp=(s,r)=>{var m;if(s.length===0)return r.classGroupId;const c=s[0],u=r.nextPart.get(c),f=u?Tp(s.slice(1),u):void 0;if(f)return f;if(r.validators.length===0)return;const d=s.join(B_);return(m=r.validators.find(({validator:g})=>g(d)))==null?void 0:m.classGroupId},N0=/^\[(.+)\]$/,qb=s=>{if(N0.test(s)){const r=N0.exec(s)[1],c=r==null?void 0:r.substring(0,r.indexOf(":"));if(c)return"arbitrary.."+c}},Ib=s=>{const{theme:r,prefix:c}=s,u={nextPart:new Map,validators:[]};return Qb(Object.entries(s.classGroups),c).forEach(([d,m])=>{M_(m,u,d,r)}),u},M_=(s,r,c,u)=>{s.forEach(f=>{if(typeof f=="string"){const d=f===""?r:M0(r,f);d.classGroupId=c;return}if(typeof f=="function"){if(Xb(f)){M_(f(u),r,c,u);return}r.validators.push({validator:f,classGroupId:c});return}Object.entries(f).forEach(([d,m])=>{M_(m,M0(r,d),c,u)})})},M0=(s,r)=>{let c=s;return r.split(B_).forEach(u=>{c.nextPart.has(u)||c.nextPart.set(u,{nextPart:new Map,validators:[]}),c=c.nextPart.get(u)}),c},Xb=s=>s.isThemeGetter,Qb=(s,r)=>r?s.map(([c,u])=>{const f=u.map(d=>typeof d=="string"?r+d:typeof d=="object"?Object.fromEntries(Object.entries(d).map(([m,g])=>[r+m,g])):d);return[c,f]}):s,Gb=s=>{if(s<1)return{get:()=>{},set:()=>{}};let r=0,c=new Map,u=new Map;const f=(d,m)=>{c.set(d,m),r++,r>s&&(r=0,u=c,c=new Map)};return{get(d){let m=c.get(d);if(m!==void 0)return m;if((m=u.get(d))!==void 0)return f(d,m),m},set(d,m){c.has(d)?c.set(d,m):f(d,m)}}},Rp="!",Vb=s=>{const{separator:r,experimentalParseClassName:c}=s,u=r.length===1,f=r[0],d=r.length,m=g=>{const p=[];let j=0,k=0,b;for(let B=0;B<g.length;B++){let F=g[B];if(j===0){if(F===f&&(u||g.slice(B,B+d)===r)){p.push(g.slice(k,B)),k=B+d;continue}if(F==="/"){b=B;continue}}F==="["?j++:F==="]"&&j--}const v=p.length===0?g:g.substring(k),z=v.startsWith(Rp),E=z?v.substring(1):v,$=b&&b>k?b-k:void 0;return{modifiers:p,hasImportantModifier:z,baseClassName:E,maybePostfixModifierPosition:$}};return c?g=>c({className:g,parseClassName:m}):m},Fb=s=>{if(s.length<=1)return s;const r=[];let c=[];return s.forEach(u=>{u[0]==="["?(r.push(...c.sort(),u),c=[]):c.push(u)}),r.push(...c.sort()),r},Wb=s=>({cache:Gb(s.cacheSize),parseClassName:Vb(s),...Yb(s)}),Kb=/\s+/,Zb=(s,r)=>{const{parseClassName:c,getClassGroupId:u,getConflictingClassGroupIds:f}=r,d=[],m=s.trim().split(Kb);let g="";for(let p=m.length-1;p>=0;p-=1){const j=m[p],{modifiers:k,hasImportantModifier:b,baseClassName:v,maybePostfixModifierPosition:z}=c(j);let E=!!z,$=u(E?v.substring(0,z):v);if(!$){if(!E){g=j+(g.length>0?" "+g:g);continue}if($=u(v),!$){g=j+(g.length>0?" "+g:g);continue}E=!1}const B=Fb(k).join(":"),F=b?B+Rp:B,Me=F+$;if(d.includes(Me))continue;d.push(Me);const te=f($,E);for(let C=0;C<te.length;++C){const ie=te[C];d.push(F+ie)}g=j+(g.length>0?" "+g:g)}return g};function Jb(){let s=0,r,c,u="";for(;s<arguments.length;)(r=arguments[s++])&&(c=zp(r))&&(u&&(u+=" "),u+=c);return u}const zp=s=>{if(typeof s=="string")return s;let r,c="";for(let u=0;u<s.length;u++)s[u]&&(r=zp(s[u]))&&(c&&(c+=" "),c+=r);return c};function Pb(s,...r){let c,u,f,d=m;function m(p){const j=r.reduce((k,b)=>b(k),s());return c=Wb(j),u=c.cache.get,f=c.cache.set,d=g,g(p)}function g(p){const j=u(p);if(j)return j;const k=Zb(p,c);return f(p,k),k}return function(){return d(Jb.apply(null,arguments))}}const _n=s=>{const r=c=>c[s]||[];return r.isThemeGetter=!0,r},Ap=/^\[(?:([a-z-]+):)?(.+)\]$/i,e1=/^\d+\/\d+$/,t1=new Set(["px","full","screen"]),n1=/^(\d+(\.\d+)?)?(xs|sm|md|lg|xl)$/,l1=/\d+(%|px|r?em|[sdl]?v([hwib]|min|max)|pt|pc|in|cm|mm|cap|ch|ex|r?lh|cq(w|h|i|b|min|max))|\b(calc|min|max|clamp)\(.+\)|^0$/,o1=/^(rgba?|hsla?|hwb|(ok)?(lab|lch)|color-mix)\(.+\)$/,s1=/^(inset_)?-?((\d+)?\.?(\d+)[a-z]+|0)_-?((\d+)?\.?(\d+)[a-z]+|0)/,a1=/^(url|image|image-set|cross-fade|element|(repeating-)?(linear|radial|conic)-gradient)\(.+\)$/,rs=s=>Yr(s)||t1.has(s)||e1.test(s),Ws=s=>Xr(s,"length",m1),Yr=s=>!!s&&!Number.isNaN(Number(s)),m_=s=>Xr(s,"number",Yr),ec=s=>!!s&&Number.isInteger(Number(s)),r1=s=>s.endsWith("%")&&Yr(s.slice(0,-1)),gt=s=>Ap.test(s),Ks=s=>n1.test(s),i1=new Set(["length","size","percentage"]),c1=s=>Xr(s,i1,Op),u1=s=>Xr(s,"position",Op),d1=new Set(["image","url"]),f1=s=>Xr(s,d1,g1),_1=s=>Xr(s,"",h1),tc=()=>!0,Xr=(s,r,c)=>{const u=Ap.exec(s);return u?u[1]?typeof r=="string"?u[1]===r:r.has(u[1]):c(u[2]):!1},m1=s=>l1.test(s)&&!o1.test(s),Op=()=>!1,h1=s=>s1.test(s),g1=s=>a1.test(s),p1=()=>{const s=_n("colors"),r=_n("spacing"),c=_n("blur"),u=_n("brightness"),f=_n("borderColor"),d=_n("borderRadius"),m=_n("borderSpacing"),g=_n("borderWidth"),p=_n("contrast"),j=_n("grayscale"),k=_n("hueRotate"),b=_n("invert"),v=_n("gap"),z=_n("gradientColorStops"),E=_n("gradientColorStopPositions"),$=_n("inset"),B=_n("margin"),F=_n("opacity"),Me=_n("padding"),te=_n("saturate"),C=_n("scale"),ie=_n("sepia"),de=_n("skew"),W=_n("space"),Ee=_n("translate"),ne=()=>["auto","contain","none"],Ce=()=>["auto","hidden","clip","visible","scroll"],pe=()=>["auto",gt,r],V=()=>[gt,r],Ie=()=>["",rs,Ws],he=()=>["auto",Yr,gt],ue=()=>["bottom","center","left","left-bottom","left-top","right","right-bottom","right-top","top"],G=()=>["solid","dashed","dotted","double","none"],we=()=>["normal","multiply","screen","overlay","darken","lighten","color-dodge","color-burn","hard-light","soft-light","difference","exclusion","hue","saturation","color","luminosity"],fe=()=>["start","end","center","between","around","evenly","stretch"],Oe=()=>["","0",gt],Se=()=>["auto","avoid","all","avoid-page","page","left","right","column"],Qe=()=>[Yr,gt];return{cacheSize:500,separator:":",theme:{colors:[tc],spacing:[rs,Ws],blur:["none","",Ks,gt],brightness:Qe(),borderColor:[s],borderRadius:["none","","full",Ks,gt],borderSpacing:V(),borderWidth:Ie(),contrast:Qe(),grayscale:Oe(),hueRotate:Qe(),invert:Oe(),gap:V(),gradientColorStops:[s],gradientColorStopPositions:[r1,Ws],inset:pe(),margin:pe(),opacity:Qe(),padding:V(),saturate:Qe(),scale:Qe(),sepia:Oe(),skew:Qe(),space:V(),translate:V()},classGroups:{aspect:[{aspect:["auto","square","video",gt]}],container:["container"],columns:[{columns:[Ks]}],"break-after":[{"break-after":Se()}],"break-before":[{"break-before":Se()}],"break-inside":[{"break-inside":["auto","avoid","avoid-page","avoid-column"]}],"box-decoration":[{"box-decoration":["slice","clone"]}],box:[{box:["border","content"]}],display:["block","inline-block","inline","flex","inline-flex","table","inline-table","table-caption","table-cell","table-column","table-column-group","table-footer-group","table-header-group","table-row-group","table-row","flow-root","grid","inline-grid","contents","list-item","hidden"],float:[{float:["right","left","none","start","end"]}],clear:[{clear:["left","right","both","none","start","end"]}],isolation:["isolate","isolation-auto"],"object-fit":[{object:["contain","cover","fill","none","scale-down"]}],"object-position":[{object:[...ue(),gt]}],overflow:[{overflow:Ce()}],"overflow-x":[{"overflow-x":Ce()}],"overflow-y":[{"overflow-y":Ce()}],overscroll:[{overscroll:ne()}],"overscroll-x":[{"overscroll-x":ne()}],"overscroll-y":[{"overscroll-y":ne()}],position:["static","fixed","absolute","relative","sticky"],inset:[{inset:[$]}],"inset-x":[{"inset-x":[$]}],"inset-y":[{"inset-y":[$]}],start:[{start:[$]}],end:[{end:[$]}],top:[{top:[$]}],right:[{right:[$]}],bottom:[{bottom:[$]}],left:[{left:[$]}],visibility:["visible","invisible","collapse"],z:[{z:["auto",ec,gt]}],basis:[{basis:pe()}],"flex-direction":[{flex:["row","row-reverse","col","col-reverse"]}],"flex-wrap":[{flex:["wrap","wrap-reverse","nowrap"]}],flex:[{flex:["1","auto","initial","none",gt]}],grow:[{grow:Oe()}],shrink:[{shrink:Oe()}],order:[{order:["first","last","none",ec,gt]}],"grid-cols":[{"grid-cols":[tc]}],"col-start-end":[{col:["auto",{span:["full",ec,gt]},gt]}],"col-start":[{"col-start":he()}],"col-end":[{"col-end":he()}],"grid-rows":[{"grid-rows":[tc]}],"row-start-end":[{row:["auto",{span:[ec,gt]},gt]}],"row-start":[{"row-start":he()}],"row-end":[{"row-end":he()}],"grid-flow":[{"grid-flow":["row","col","dense","row-dense","col-dense"]}],"auto-cols":[{"auto-cols":["auto","min","max","fr",gt]}],"auto-rows":[{"auto-rows":["auto","min","max","fr",gt]}],gap:[{gap:[v]}],"gap-x":[{"gap-x":[v]}],"gap-y":[{"gap-y":[v]}],"justify-content":[{justify:["normal",...fe()]}],"justify-items":[{"justify-items":["start","end","center","stretch"]}],"justify-self":[{"justify-self":["auto","start","end","center","stretch"]}],"align-content":[{content:["normal",...fe(),"baseline"]}],"align-items":[{items:["start","end","center","baseline","stretch"]}],"align-self":[{self:["auto","start","end","center","stretch","baseline"]}],"place-content":[{"place-content":[...fe(),"baseline"]}],"place-items":[{"place-items":["start","end","center","baseline","stretch"]}],"place-self":[{"place-self":["auto","start","end","center","stretch"]}],p:[{p:[Me]}],px:[{px:[Me]}],py:[{py:[Me]}],ps:[{ps:[Me]}],pe:[{pe:[Me]}],pt:[{pt:[Me]}],pr:[{pr:[Me]}],pb:[{pb:[Me]}],pl:[{pl:[Me]}],m:[{m:[B]}],mx:[{mx:[B]}],my:[{my:[B]}],ms:[{ms:[B]}],me:[{me:[B]}],mt:[{mt:[B]}],mr:[{mr:[B]}],mb:[{mb:[B]}],ml:[{ml:[B]}],"space-x":[{"space-x":[W]}],"space-x-reverse":["space-x-reverse"],"space-y":[{"space-y":[W]}],"space-y-reverse":["space-y-reverse"],w:[{w:["auto","min","max","fit","svw","lvw","dvw",gt,r]}],"min-w":[{"min-w":[gt,r,"min","max","fit"]}],"max-w":[{"max-w":[gt,r,"none","full","min","max","fit","prose",{screen:[Ks]},Ks]}],h:[{h:[gt,r,"auto","min","max","fit","svh","lvh","dvh"]}],"min-h":[{"min-h":[gt,r,"min","max","fit","svh","lvh","dvh"]}],"max-h":[{"max-h":[gt,r,"min","max","fit","svh","lvh","dvh"]}],size:[{size:[gt,r,"auto","min","max","fit"]}],"font-size":[{text:["base",Ks,Ws]}],"font-smoothing":["antialiased","subpixel-antialiased"],"font-style":["italic","not-italic"],"font-weight":[{font:["thin","extralight","light","normal","medium","semibold","bold","extrabold","black",m_]}],"font-family":[{font:[tc]}],"fvn-normal":["normal-nums"],"fvn-ordinal":["ordinal"],"fvn-slashed-zero":["slashed-zero"],"fvn-figure":["lining-nums","oldstyle-nums"],"fvn-spacing":["proportional-nums","tabular-nums"],"fvn-fraction":["diagonal-fractions","stacked-fractions"],tracking:[{tracking:["tighter","tight","normal","wide","wider","widest",gt]}],"line-clamp":[{"line-clamp":["none",Yr,m_]}],leading:[{leading:["none","tight","snug","normal","relaxed","loose",rs,gt]}],"list-image":[{"list-image":["none",gt]}],"list-style-type":[{list:["none","disc","decimal",gt]}],"list-style-position":[{list:["inside","outside"]}],"placeholder-color":[{placeholder:[s]}],"placeholder-opacity":[{"placeholder-opacity":[F]}],"text-alignment":[{text:["left","center","right","justify","start","end"]}],"text-color":[{text:[s]}],"text-opacity":[{"text-opacity":[F]}],"text-decoration":["underline","overline","line-through","no-underline"],"text-decoration-style":[{decoration:[...G(),"wavy"]}],"text-decoration-thickness":[{decoration:["auto","from-font",rs,Ws]}],"underline-offset":[{"underline-offset":["auto",rs,gt]}],"text-decoration-color":[{decoration:[s]}],"text-transform":["uppercase","lowercase","capitalize","normal-case"],"text-overflow":["truncate","text-ellipsis","text-clip"],"text-wrap":[{text:["wrap","nowrap","balance","pretty"]}],indent:[{indent:V()}],"vertical-align":[{align:["baseline","top","middle","bottom","text-top","text-bottom","sub","super",gt]}],whitespace:[{whitespace:["normal","nowrap","pre","pre-line","pre-wrap","break-spaces"]}],break:[{break:["normal","words","all","keep"]}],hyphens:[{hyphens:["none","manual","auto"]}],content:[{content:["none",gt]}],"bg-attachment":[{bg:["fixed","local","scroll"]}],"bg-clip":[{"bg-clip":["border","padding","content","text"]}],"bg-opacity":[{"bg-opacity":[F]}],"bg-origin":[{"bg-origin":["border","padding","content"]}],"bg-position":[{bg:[...ue(),u1]}],"bg-repeat":[{bg:["no-repeat",{repeat:["","x","y","round","space"]}]}],"bg-size":[{bg:["auto","cover","contain",c1]}],"bg-image":[{bg:["none",{"gradient-to":["t","tr","r","br","b","bl","l","tl"]},f1]}],"bg-color":[{bg:[s]}],"gradient-from-pos":[{from:[E]}],"gradient-via-pos":[{via:[E]}],"gradient-to-pos":[{to:[E]}],"gradient-from":[{from:[z]}],"gradient-via":[{via:[z]}],"gradient-to":[{to:[z]}],rounded:[{rounded:[d]}],"rounded-s":[{"rounded-s":[d]}],"rounded-e":[{"rounded-e":[d]}],"rounded-t":[{"rounded-t":[d]}],"rounded-r":[{"rounded-r":[d]}],"rounded-b":[{"rounded-b":[d]}],"rounded-l":[{"rounded-l":[d]}],"rounded-ss":[{"rounded-ss":[d]}],"rounded-se":[{"rounded-se":[d]}],"rounded-ee":[{"rounded-ee":[d]}],"rounded-es":[{"rounded-es":[d]}],"rounded-tl":[{"rounded-tl":[d]}],"rounded-tr":[{"rounded-tr":[d]}],"rounded-br":[{"rounded-br":[d]}],"rounded-bl":[{"rounded-bl":[d]}],"border-w":[{border:[g]}],"border-w-x":[{"border-x":[g]}],"border-w-y":[{"border-y":[g]}],"border-w-s":[{"border-s":[g]}],"border-w-e":[{"border-e":[g]}],"border-w-t":[{"border-t":[g]}],"border-w-r":[{"border-r":[g]}],"border-w-b":[{"border-b":[g]}],"border-w-l":[{"border-l":[g]}],"border-opacity":[{"border-opacity":[F]}],"border-style":[{border:[...G(),"hidden"]}],"divide-x":[{"divide-x":[g]}],"divide-x-reverse":["divide-x-reverse"],"divide-y":[{"divide-y":[g]}],"divide-y-reverse":["divide-y-reverse"],"divide-opacity":[{"divide-opacity":[F]}],"divide-style":[{divide:G()}],"border-color":[{border:[f]}],"border-color-x":[{"border-x":[f]}],"border-color-y":[{"border-y":[f]}],"border-color-s":[{"border-s":[f]}],"border-color-e":[{"border-e":[f]}],"border-color-t":[{"border-t":[f]}],"border-color-r":[{"border-r":[f]}],"border-color-b":[{"border-b":[f]}],"border-color-l":[{"border-l":[f]}],"divide-color":[{divide:[f]}],"outline-style":[{outline:["",...G()]}],"outline-offset":[{"outline-offset":[rs,gt]}],"outline-w":[{outline:[rs,Ws]}],"outline-color":[{outline:[s]}],"ring-w":[{ring:Ie()}],"ring-w-inset":["ring-inset"],"ring-color":[{ring:[s]}],"ring-opacity":[{"ring-opacity":[F]}],"ring-offset-w":[{"ring-offset":[rs,Ws]}],"ring-offset-color":[{"ring-offset":[s]}],shadow:[{shadow:["","inner","none",Ks,_1]}],"shadow-color":[{shadow:[tc]}],opacity:[{opacity:[F]}],"mix-blend":[{"mix-blend":[...we(),"plus-lighter","plus-darker"]}],"bg-blend":[{"bg-blend":we()}],filter:[{filter:["","none"]}],blur:[{blur:[c]}],brightness:[{brightness:[u]}],contrast:[{contrast:[p]}],"drop-shadow":[{"drop-shadow":["","none",Ks,gt]}],grayscale:[{grayscale:[j]}],"hue-rotate":[{"hue-rotate":[k]}],invert:[{invert:[b]}],saturate:[{saturate:[te]}],sepia:[{sepia:[ie]}],"backdrop-filter":[{"backdrop-filter":["","none"]}],"backdrop-blur":[{"backdrop-blur":[c]}],"backdrop-brightness":[{"backdrop-brightness":[u]}],"backdrop-contrast":[{"backdrop-contrast":[p]}],"backdrop-grayscale":[{"backdrop-grayscale":[j]}],"backdrop-hue-rotate":[{"backdrop-hue-rotate":[k]}],"backdrop-invert":[{"backdrop-invert":[b]}],"backdrop-opacity":[{"backdrop-opacity":[F]}],"backdrop-saturate":[{"backdrop-saturate":[te]}],"backdrop-sepia":[{"backdrop-sepia":[ie]}],"border-collapse":[{border:["collapse","separate"]}],"border-spacing":[{"border-spacing":[m]}],"border-spacing-x":[{"border-spacing-x":[m]}],"border-spacing-y":[{"border-spacing-y":[m]}],"table-layout":[{table:["auto","fixed"]}],caption:[{caption:["top","bottom"]}],transition:[{transition:["none","all","","colors","opacity","shadow","transform",gt]}],duration:[{duration:Qe()}],ease:[{ease:["linear","in","out","in-out",gt]}],delay:[{delay:Qe()}],animate:[{animate:["none","spin","ping","pulse","bounce",gt]}],transform:[{transform:["","gpu","none"]}],scale:[{scale:[C]}],"scale-x":[{"scale-x":[C]}],"scale-y":[{"scale-y":[C]}],rotate:[{rotate:[ec,gt]}],"translate-x":[{"translate-x":[Ee]}],"translate-y":[{"translate-y":[Ee]}],"skew-x":[{"skew-x":[de]}],"skew-y":[{"skew-y":[de]}],"transform-origin":[{origin:["center","top","top-right","right","bottom-right","bottom","bottom-left","left","top-left",gt]}],accent:[{accent:["auto",s]}],appearance:[{appearance:["none","auto"]}],cursor:[{cursor:["auto","default","pointer","wait","text","move","help","not-allowed","none","context-menu","progress","cell","crosshair","vertical-text","alias","copy","no-drop","grab","grabbing","all-scroll","col-resize","row-resize","n-resize","e-resize","s-resize","w-resize","ne-resize","nw-resize","se-resize","sw-resize","ew-resize","ns-resize","nesw-resize","nwse-resize","zoom-in","zoom-out",gt]}],"caret-color":[{caret:[s]}],"pointer-events":[{"pointer-events":["none","auto"]}],resize:[{resize:["none","y","x",""]}],"scroll-behavior":[{scroll:["auto","smooth"]}],"scroll-m":[{"scroll-m":V()}],"scroll-mx":[{"scroll-mx":V()}],"scroll-my":[{"scroll-my":V()}],"scroll-ms":[{"scroll-ms":V()}],"scroll-me":[{"scroll-me":V()}],"scroll-mt":[{"scroll-mt":V()}],"scroll-mr":[{"scroll-mr":V()}],"scroll-mb":[{"scroll-mb":V()}],"scroll-ml":[{"scroll-ml":V()}],"scroll-p":[{"scroll-p":V()}],"scroll-px":[{"scroll-px":V()}],"scroll-py":[{"scroll-py":V()}],"scroll-ps":[{"scroll-ps":V()}],"scroll-pe":[{"scroll-pe":V()}],"scroll-pt":[{"scroll-pt":V()}],"scroll-pr":[{"scroll-pr":V()}],"scroll-pb":[{"scroll-pb":V()}],"scroll-pl":[{"scroll-pl":V()}],"snap-align":[{snap:["start","end","center","align-none"]}],"snap-stop":[{snap:["normal","always"]}],"snap-type":[{snap:["none","x","y","both"]}],"snap-strictness":[{snap:["mandatory","proximity"]}],touch:[{touch:["auto","none","manipulation"]}],"touch-x":[{"touch-pan":["x","left","right"]}],"touch-y":[{"touch-pan":["y","up","down"]}],"touch-pz":["touch-pinch-zoom"],select:[{select:["none","text","all","auto"]}],"will-change":[{"will-change":["auto","scroll","contents","transform",gt]}],fill:[{fill:[s,"none"]}],"stroke-w":[{stroke:[rs,Ws,m_]}],stroke:[{stroke:[s,"none"]}],sr:["sr-only","not-sr-only"],"forced-color-adjust":[{"forced-color-adjust":["auto","none"]}]},conflictingClassGroups:{overflow:["overflow-x","overflow-y"],overscroll:["overscroll-x","overscroll-y"],inset:["inset-x","inset-y","start","end","top","right","bottom","left"],"inset-x":["right","left"],"inset-y":["top","bottom"],flex:["basis","grow","shrink"],gap:["gap-x","gap-y"],p:["px","py","ps","pe","pt","pr","pb","pl"],px:["pr","pl"],py:["pt","pb"],m:["mx","my","ms","me","mt","mr","mb","ml"],mx:["mr","ml"],my:["mt","mb"],size:["w","h"],"font-size":["leading"],"fvn-normal":["fvn-ordinal","fvn-slashed-zero","fvn-figure","fvn-spacing","fvn-fraction"],"fvn-ordinal":["fvn-normal"],"fvn-slashed-zero":["fvn-normal"],"fvn-figure":["fvn-normal"],"fvn-spacing":["fvn-normal"],"fvn-fraction":["fvn-normal"],"line-clamp":["display","overflow"],rounded:["rounded-s","rounded-e","rounded-t","rounded-r","rounded-b","rounded-l","rounded-ss","rounded-se","rounded-ee","rounded-es","rounded-tl","rounded-tr","rounded-br","rounded-bl"],"rounded-s":["rounded-ss","rounded-es"],"rounded-e":["rounded-se","rounded-ee"],"rounded-t":["rounded-tl","rounded-tr"],"rounded-r":["rounded-tr","rounded-br"],"rounded-b":["rounded-br","rounded-bl"],"rounded-l":["rounded-tl","rounded-bl"],"border-spacing":["border-spacing-x","border-spacing-y"],"border-w":["border-w-s","border-w-e","border-w-t","border-w-r","border-w-b","border-w-l"],"border-w-x":["border-w-r","border-w-l"],"border-w-y":["border-w-t","border-w-b"],"border-color":["border-color-s","border-color-e","border-color-t","border-color-r","border-color-b","border-color-l"],"border-color-x":["border-color-r","border-color-l"],"border-color-y":["border-color-t","border-color-b"],"scroll-m":["scroll-mx","scroll-my","scroll-ms","scroll-me","scroll-mt","scroll-mr","scroll-mb","scroll-ml"],"scroll-mx":["scroll-mr","scroll-ml"],"scroll-my":["scroll-mt","scroll-mb"],"scroll-p":["scroll-px","scroll-py","scroll-ps","scroll-pe","scroll-pt","scroll-pr","scroll-pb","scroll-pl"],"scroll-px":["scroll-pr","scroll-pl"],"scroll-py":["scroll-pt","scroll-pb"],touch:["touch-x","touch-y","touch-pz"],"touch-x":["touch"],"touch-y":["touch"],"touch-pz":["touch"]},conflictingClassGroupModifiers:{"font-size":["leading"]}}},y1=Pb(p1);function Pt(...s){return y1(Ep(s))}function x1({className:s,variant:r="dart",...c}){const u=h.useId();switch(r){case"core-void":return o.jsxs("svg",{viewBox:"0 0 100 100",fill:"none",xmlns:"http://www.w3.org/2000/svg",className:Pt("shrink-0 select-none",s),"aria-label":"AIM Logo: Core Void",...c,children:[o.jsxs("mask",{id:`void-mask-${u}`,children:[o.jsx("rect",{width:"100",height:"100",fill:"#ffffff"}),o.jsx("line",{x1:"50",y1:"10",x2:"50",y2:"28",stroke:"#000000",strokeWidth:"5",strokeLinecap:"round"})]}),o.jsx("circle",{cx:"50",cy:"50",r:"34",stroke:"currentColor",strokeWidth:"7",mask:`url(#void-mask-${u})`}),o.jsx("circle",{cx:"50",cy:"50",r:"14",fill:"currentColor"})]});case"nexus":return o.jsxs("svg",{viewBox:"0 0 100 100",fill:"none",xmlns:"http://www.w3.org/2000/svg",className:Pt("shrink-0 select-none",s),"aria-label":"AIM Logo: Tri-Agent Nexus",...c,children:[o.jsx("line",{x1:"50",y1:"24",x2:"50",y2:"50",stroke:"currentColor",strokeWidth:"3",strokeLinecap:"round"}),o.jsx("line",{x1:"26",y1:"68",x2:"50",y2:"50",stroke:"currentColor",strokeWidth:"3",strokeLinecap:"round"}),o.jsx("line",{x1:"74",y1:"68",x2:"50",y2:"50",stroke:"currentColor",strokeWidth:"3",strokeLinecap:"round"}),o.jsx("circle",{cx:"50",cy:"24",r:"8",fill:"currentColor"}),o.jsx("circle",{cx:"26",cy:"68",r:"7",fill:"currentColor"}),o.jsx("circle",{cx:"74",cy:"68",r:"7",fill:"currentColor"}),o.jsx("circle",{cx:"50",cy:"50",r:"5",stroke:"currentColor",strokeWidth:"2.5",fill:"var(--background, #000)"})]});case"monogram":return o.jsxs("svg",{viewBox:"0 0 100 100",fill:"none",xmlns:"http://www.w3.org/2000/svg",className:Pt("shrink-0 select-none",s),"aria-label":"AIM Logo: Architectural Monogram",...c,children:[o.jsx("path",{d:"M22 74L50 22L78 74",stroke:"currentColor",strokeWidth:"7.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("line",{x1:"32",y1:"56",x2:"68",y2:"56",stroke:"currentColor",strokeWidth:"5.5",strokeLinecap:"round"})]});case"aperture":return o.jsxs("svg",{viewBox:"0 0 100 100",fill:"none",xmlns:"http://www.w3.org/2000/svg",className:Pt("shrink-0 select-none",s),"aria-label":"AIM Logo: Radial Aperture",...c,children:[o.jsxs("g",{stroke:"currentColor",strokeWidth:"5",strokeLinecap:"round",children:[o.jsx("line",{x1:"50",y1:"16",x2:"66",y2:"34"}),o.jsx("line",{x1:"79",y1:"33",x2:"71",y2:"57"}),o.jsx("line",{x1:"79",y1:"67",x2:"55",y2:"73"}),o.jsx("line",{x1:"50",y1:"84",x2:"34",y2:"66"}),o.jsx("line",{x1:"21",y1:"67",x2:"29",y2:"43"}),o.jsx("line",{x1:"21",y1:"33",x2:"45",y2:"27"})]}),o.jsx("circle",{cx:"50",cy:"50",r:"6",fill:"currentColor"})]});case"sandboxes":return o.jsxs("svg",{viewBox:"0 0 100 100",fill:"none",xmlns:"http://www.w3.org/2000/svg",className:Pt("shrink-0 select-none",s),"aria-label":"AIM Logo: Dual Sandboxes",...c,children:[o.jsx("circle",{cx:"39",cy:"50",r:"23",stroke:"currentColor",strokeWidth:"5.5"}),o.jsx("circle",{cx:"61",cy:"50",r:"23",stroke:"currentColor",strokeWidth:"5.5"}),o.jsx("circle",{cx:"50",cy:"50",r:"4.5",fill:"currentColor"})]});case"dart":return o.jsxs("svg",{viewBox:"0 0 100 100",fill:"none",xmlns:"http://www.w3.org/2000/svg",className:Pt("shrink-0 select-none",s),"aria-label":"AIM Logo: Forward Dart",...c,children:[o.jsx("path",{d:"M78 22L28 42L48 52L58 72L78 22Z",fill:"currentColor"}),o.jsx("circle",{cx:"28",cy:"72",r:"5",fill:"currentColor"})]})}}function b1({currentView:s,selectedProfile:r,profiles:c,daemonInfo:u,version:f="v0.12.0",isOnline:d=!0,onSelectView:m,onSelectProfile:g,onOpenAddProfile:p,onOpenDaemonModal:j}){const[k,b]=h.useState(""),v=h.useRef(null);h.useEffect(()=>{const C=ie=>{var de,W,Ee;ie.metaKey||ie.ctrlKey||ie.altKey||document.querySelector('[role="dialog"]')||(ie.key==="f"||ie.key==="F")&&((de=document.activeElement)==null?void 0:de.tagName)!=="INPUT"&&((W=document.activeElement)==null?void 0:W.tagName)!=="TEXTAREA"&&(ie.preventDefault(),(Ee=v.current)==null||Ee.focus())};return window.addEventListener("keydown",C),()=>window.removeEventListener("keydown",C)},[]);const z=C=>{var de,W,Ee;if(!C.has_credentials)return{color:"bg-status-warning-solid",text:"text-status-warning",label:"Auth"};if((de=C.quota)!=null&&de.is_exhausted||((W=C.quota)==null?void 0:W.bottleneck_pct)!=null&&C.quota.bottleneck_pct<=0)return{color:"bg-status-danger-solid",text:"text-status-danger",label:"0%"};if(((Ee=C.quota)==null?void 0:Ee.bottleneck_pct)==null)return{color:"bg-muted-foreground/40",text:"text-muted-foreground",label:"--%"};const ie=C.quota.bottleneck_pct;return ie<=30?{color:"bg-status-warning-solid",text:"text-status-warning",label:`${ie}%`}:{color:"bg-status-success-solid",text:"text-status-success",label:`${ie}%`}},E=h.useMemo(()=>c.filter(C=>!C.is_leased&&!C.name.includes(":")),[c]),$=h.useMemo(()=>{if(!k.trim())return E;const C=k.toLowerCase().trim();return E.filter(ie=>ie.name.toLowerCase().includes(C)||ie.agent.toLowerCase().includes(C)||ie.adapters&&ie.adapters.some(de=>de.agent.toLowerCase().includes(C)))},[E,k]),B=s==="overview",F=s==="sessions"||s==="session-detail",Me=s==="mcp",te=s==="network";return o.jsxs("aside",{className:"w-64 border-r border-border bg-background flex flex-col h-full shrink-0 select-none",children:[o.jsx("div",{className:"h-14 px-4 border-b border-border flex items-center justify-between",children:o.jsxs("button",{type:"button",onClick:()=>m("overview"),className:"flex items-center gap-2.5 text-left hover:opacity-85 transition-opacity cursor-pointer group",title:"AIM Home",children:[o.jsx("div",{className:"h-7 w-7 rounded-[6px] bg-card border border-border flex items-center justify-center shadow-xs",children:o.jsx(x1,{className:"h-4 w-4 text-foreground"})}),o.jsxs("div",{className:"flex items-center gap-1.5",children:[o.jsx("span",{className:"font-semibold text-sm tracking-tight text-foreground",children:"AIM"}),o.jsx("span",{className:"sr-only",children:"Agent Identity Manager"}),o.jsx("span",{className:"text-xs font-mono px-1.5 py-0.5 rounded border border-border bg-card text-muted-foreground",children:f})]})]})}),o.jsx("div",{className:"px-3 pt-3 pb-1",children:o.jsxs("div",{className:"relative flex items-center",children:[o.jsx(ub,{className:"absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/80 pointer-events-none"}),o.jsx("input",{ref:v,type:"text",value:k,onChange:C=>b(C.target.value),placeholder:"Find","aria-label":"Filter navigation items and profiles",className:"w-full h-7 pl-8 pr-7 text-xs font-sans bg-[#0c0c0c] border border-border/80 focus:border-border text-foreground placeholder:text-muted-foreground rounded-geist focus:outline-none focus:ring-1 focus:ring-ring transition-colors"}),o.jsx("kbd",{className:"absolute right-2 top-1/2 -translate-y-1/2 border border-border bg-muted/60 px-1 py-[1px] rounded text-xs font-mono text-muted-foreground pointer-events-none",children:"F"})]})}),o.jsxs("div",{className:"flex-1 overflow-y-auto px-3 py-2 space-y-5",children:[o.jsxs("div",{role:"tablist","aria-label":"Views",className:"space-y-0.5",children:[o.jsxs("button",{role:"tab","aria-selected":B,type:"button",onClick:()=>m("overview"),className:`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${B?"bg-secondary text-foreground border border-border shadow-sm font-medium":"text-muted-foreground hover:text-foreground hover:bg-card border border-transparent"}`,children:[o.jsx(db,{className:`h-3.5 w-3.5 ${B?"text-foreground":"text-muted-foreground"}`}),o.jsx("span",{children:"Profiles"}),o.jsx("span",{className:"ml-auto text-xs font-mono text-muted-foreground/80",children:c.length})]}),o.jsxs("button",{role:"tab","aria-selected":F,type:"button",onClick:()=>m("sessions"),className:`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${F?"bg-secondary text-foreground border border-border shadow-sm font-medium":"text-muted-foreground hover:text-foreground hover:bg-card border border-transparent"}`,children:[o.jsx(kp,{className:`h-3.5 w-3.5 ${F?"text-foreground":"text-muted-foreground"}`}),o.jsx("span",{children:"Sessions"})]}),o.jsxs("button",{role:"tab","aria-selected":Me,type:"button",onClick:()=>m("mcp"),className:`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${Me?"bg-secondary text-foreground border border-border shadow-sm font-medium":"text-muted-foreground hover:text-foreground hover:bg-card border border-transparent"}`,children:[o.jsx(fb,{className:`h-3.5 w-3.5 ${Me?"text-foreground":"text-muted-foreground"}`}),o.jsx("span",{children:"MCP Catalog"})]}),o.jsxs("button",{role:"tab","aria-selected":te,type:"button",onClick:()=>m("network"),className:`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${te?"bg-secondary text-foreground border border-border shadow-sm font-medium":"text-muted-foreground hover:text-foreground hover:bg-card border border-transparent"}`,children:[o.jsx(x0,{className:`h-3.5 w-3.5 ${te?"text-accent-cyan":"text-muted-foreground"}`}),o.jsx("span",{children:"P2P Network"})]})]}),o.jsxs("div",{className:"space-y-1",children:[o.jsxs("div",{className:"flex items-center justify-between px-2 pb-1",children:[o.jsxs("span",{className:"text-xs font-mono uppercase tracking-wider text-muted-foreground/80",children:["Active Profiles (",$.length,")"]}),o.jsx("button",{type:"button",onClick:p,title:"Add new profile","aria-label":"Add new profile",className:"h-5 w-5 rounded-[4px] flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer",children:o.jsx(_b,{className:"h-3.5 w-3.5"})})]}),o.jsx("div",{className:"space-y-0.5",children:$.length===0?o.jsx("div",{className:"px-2.5 py-2 text-xs font-sans text-muted-foreground/80 italic",children:"No matching profiles"}):$.map(C=>{const ie=z(C),de=s==="profile-detail"&&r===C.name;return o.jsxs("button",{type:"button",onClick:()=>g(C.name),className:`w-full flex items-center justify-between px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left group ${de?"bg-secondary text-foreground border border-border shadow-sm font-medium":"text-muted-foreground hover:text-foreground hover:bg-card border border-transparent"}`,children:[o.jsxs("div",{className:"flex items-center gap-2 truncate",children:[o.jsx("span",{className:`h-1.5 w-1.5 rounded-full shrink-0 ${ie.color}`}),o.jsx("span",{className:"truncate",children:C.name})]}),o.jsx("span",{className:`text-xs font-mono tabular-nums shrink-0 ${ie.text}`,children:ie.label})]},C.name)})})]})]}),o.jsx("div",{className:"p-3 border-t border-border bg-muted/20",children:o.jsxs("button",{type:"button",onClick:j,className:"w-full flex items-center justify-between px-2.5 py-1.5 rounded-geist border border-border bg-card hover:bg-secondary hover:border-border/80 transition-colors cursor-pointer text-xs group",title:"AIM Background Daemon status (click to manage)",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx(x0,{className:"h-3.5 w-3.5 text-muted-foreground/60 group-hover:text-foreground transition-colors shrink-0"}),o.jsxs("span",{className:"font-mono text-xs text-muted-foreground group-hover:text-foreground transition-colors",children:["Daemon: ",u!=null&&u.active?"Active":u!=null&&u.installed?"Inactive":"Off"]})]}),o.jsxs("div",{className:"flex items-center gap-1.5 px-1.5 py-0.5 rounded-[4px] border border-border/60 bg-muted/40 text-xs font-mono text-muted-foreground shrink-0",title:d?"Backend connected":"Backend offline",children:[o.jsxs("span",{className:"relative flex h-1.5 w-1.5 items-center justify-center",children:[d&&o.jsx("span",{className:"motion-reduce:animate-none animate-ping absolute inline-flex h-full w-full rounded-full bg-status-success-solid opacity-75"}),o.jsx("span",{className:Pt("relative inline-flex h-1.5 w-1.5 rounded-full",d?"bg-status-success-solid shadow-[0_0_8px_#30a46c]":"bg-status-danger-solid")})]}),o.jsx("span",{children:d?"live":"offline"})]})]})})]})}const Dp=sb,v1=lb,Lp=h.forwardRef(({className:s,...r},c)=>o.jsx(hp,{ref:c,className:Pt("fixed inset-0 z-50 bg-black/80 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",s),...r}));Lp.displayName=hp.displayName;const $_=h.forwardRef(({className:s,children:r,...c},u)=>o.jsxs(v1,{children:[o.jsx(Lp,{}),o.jsxs(gp,{ref:u,className:Pt("fixed left-[50%] top-[50%] z-50 flex flex-col w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border border-border bg-[#111111] p-6 shadow-2xl duration-200 rounded-geist data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]",s),...c,children:[r,o.jsxs(ob,{className:"absolute right-4 top-4 rounded-geist p-1 text-muted-foreground opacity-70 transition-opacity hover:opacity-100 hover:bg-[#1f1f1f] hover:text-foreground focus:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3] disabled:pointer-events-none",children:[o.jsx(Cp,{className:"h-4 w-4"}),o.jsx("span",{className:"sr-only",children:"Close"})]})]})]}));$_.displayName=gp.displayName;const U_=({className:s,...r})=>o.jsx("div",{className:Pt("flex flex-col space-y-1.5 text-center sm:text-left",s),...r});U_.displayName="DialogHeader";const H_=({className:s,...r})=>o.jsx("div",{className:Pt("flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",s),...r});H_.displayName="DialogFooter";const Y_=h.forwardRef(({className:s,...r},c)=>o.jsx(pp,{ref:c,className:Pt("text-lg font-semibold leading-none tracking-tight",s),...r}));Y_.displayName=pp.displayName;const q_=h.forwardRef(({className:s,...r},c)=>o.jsx(yp,{ref:c,className:Pt("text-sm text-muted-foreground",s),...r}));q_.displayName=yp.displayName;const E0=s=>typeof s=="boolean"?`${s}`:s===0?"0":s,T0=Ep,I_=(s,r)=>c=>{var u;if((r==null?void 0:r.variants)==null)return T0(s,c==null?void 0:c.class,c==null?void 0:c.className);const{variants:f,defaultVariants:d}=r,m=Object.keys(f).map(j=>{const k=c==null?void 0:c[j],b=d==null?void 0:d[j];if(k===null)return null;const v=E0(k)||E0(b);return f[j][v]}),g=c&&Object.entries(c).reduce((j,k)=>{let[b,v]=k;return v===void 0||(j[b]=v),j},{}),p=r==null||(u=r.compoundVariants)===null||u===void 0?void 0:u.reduce((j,k)=>{let{class:b,className:v,...z}=k;return Object.entries(z).every(E=>{let[$,B]=E;return Array.isArray(B)?B.includes({...d,...g}[$]):{...d,...g}[$]===B})?[...j,b,v]:j},[]);return T0(s,m,p,c==null?void 0:c.class,c==null?void 0:c.className)},w1=I_("inline-flex items-center justify-center whitespace-nowrap rounded-geist text-sm font-medium transition-all duration-150 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3] focus-visible:ring-offset-1 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 cursor-pointer",{variants:{variant:{default:"bg-[#ededed] text-[#0a0a0a] hover:bg-white active:scale-[0.98]",destructive:"bg-[#f13242] text-white hover:bg-[#e2162a] active:scale-[0.98]",outline:"border border-[#262626] bg-[#0e0e0e] text-[#ededed] hover:bg-[#171717] hover:border-[#383838] active:scale-[0.98]",secondary:"bg-[#1a1a1a] text-[#ededed] border border-[#2a2a2a] hover:bg-[#222222] active:scale-[0.98]",ghost:"text-[#888888] hover:text-[#ededed] hover:bg-[#171717]",link:"text-[#0070f3] underline-offset-4 hover:underline"},size:{default:"h-9 px-4 py-2",sm:"h-8 px-3 text-xs",lg:"h-10 px-6 text-sm",icon:"h-8 w-8"}},defaultVariants:{variant:"default",size:"default"}}),$o=h.forwardRef(({className:s,variant:r,size:c,asChild:u=!1,...f},d)=>{const m=u?ab:"button";return o.jsx(m,{className:Pt(w1({variant:r,size:c,className:s})),ref:d,...f})});$o.displayName="Button";const S1=I_("inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-colors focus:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3]",{variants:{variant:{default:"border-transparent bg-primary text-primary-foreground",secondary:"border-border bg-secondary text-secondary-foreground",destructive:"border-status-danger-border bg-status-danger-bg text-status-danger-text",outline:"border-border text-foreground",success:"border-status-success-border bg-status-success-bg text-status-success-text",warning:"border-status-warning-border bg-status-warning-bg text-status-warning-text",cursor:"border-[#f54e00]/30 bg-[#f54e00]/10 text-[#f54e00]",thinking:"border-[#dfa88f]/30 bg-[#dfa88f]/10 text-[#dfa88f]",exec:"border-[#9fc9a2]/30 bg-[#9fc9a2]/10 text-[#9fc9a2]",read:"border-[#9fbbe0]/30 bg-[#9fbbe0]/10 text-[#9fbbe0]",edit:"border-[#c0a8dd]/30 bg-[#c0a8dd]/10 text-[#c0a8dd]",done:"border-[#c08532]/30 bg-[#c08532]/10 text-[#c08532]"}},defaultVariants:{variant:"default"}});function Du({className:s,variant:r,...c}){return o.jsx("div",{className:Pt(S1({variant:r}),s),...c})}const j1=3,k1=1e6;let h_=0;function C1(){return h_=(h_+1)%Number.MAX_SAFE_INTEGER,h_.toString()}const g_=new Map,R0=s=>{if(g_.has(s))return;const r=setTimeout(()=>{g_.delete(s),uc({type:"REMOVE_TOAST",toastId:s})},k1);g_.set(s,r)},N1=(s,r)=>{switch(r.type){case"ADD_TOAST":return{...s,toasts:[r.toast,...s.toasts].slice(0,j1)};case"UPDATE_TOAST":return{...s,toasts:s.toasts.map(c=>c.id===r.toast.id?{...c,...r.toast}:c)};case"DISMISS_TOAST":{const{toastId:c}=r;return c?R0(c):s.toasts.forEach(u=>{R0(u.id)}),{...s,toasts:s.toasts.map(u=>u.id===c||c===void 0?{...u,open:!1}:u)}}case"REMOVE_TOAST":return r.toastId===void 0?{...s,toasts:[]}:{...s,toasts:s.toasts.filter(c=>c.id!==r.toastId)}}},Zu=[];let Ju={toasts:[]};function uc(s){Ju=N1(Ju,s),Zu.forEach(r=>{r(Ju)})}function M1({...s}){const r=C1(),c=f=>uc({type:"UPDATE_TOAST",toast:{...f,id:r}}),u=()=>uc({type:"DISMISS_TOAST",toastId:r});return uc({type:"ADD_TOAST",toast:{...s,id:r,open:!0,onOpenChange:f=>{f||u()}}}),{id:r,dismiss:u,update:c}}function X_(){const[s,r]=h.useState(Ju);return h.useEffect(()=>(Zu.push(r),()=>{const c=Zu.indexOf(r);c>-1&&Zu.splice(c,1)}),[s]),{...s,toast:M1,dismiss:c=>uc({type:"DISMISS_TOAST",toastId:c})}}function E1(s){const r=s.trim(),c=r.match(/^\[(.*?)\]\s*\[(INFO|WARN|WARNING|ERROR|FATAL|DEBUG)\]\s*(.*)$/i);if(c){const[,g,p,j]=c,k=p.toUpperCase(),b=k.startsWith("WARN")?"WARN":k==="ERROR"||k==="FATAL"?"ERROR":"INFO";return{timestamp:g.endsWith("UTC")?g:`${g} UTC`,level:b,message:j}}const u=r.match(/^\[(.*?)\]\s*(.*)$/);if(u){const[,g,p]=u,j=p.toLowerCase();let k="INFO";return!j.includes("0 errors")&&j.includes("error")||j.includes("fatal")||j.includes("failed")||j.includes("panic")?k="ERROR":(j.includes("warn")||j.includes("0 configured")||j.includes("timeout")||j.includes("deadline exceeded"))&&(k="WARN"),{timestamp:g.endsWith("UTC")?g:`${g} UTC`,level:k,message:p}}const f=r.match(/^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s*(.*)$/);if(f){const[,g,p]=f,j=g.replace(/\//g,"-")+" UTC",k=p.toLowerCase();let b="INFO";return k.includes("error")||k.includes("fatal")||k.includes("failed")?b="ERROR":(k.includes("warn")||k.includes("deadline exceeded")||k.includes("timeout"))&&(b="WARN"),{timestamp:j,level:b,message:p}}const d=r.toLowerCase();let m="INFO";return!d.includes("0 errors")&&d.includes("error")||d.includes("failed")?m="ERROR":(d.includes("warn")||d.includes("timeout")||d.includes("deadline exceeded"))&&(m="WARN"),{timestamp:"",level:m,message:r}}const z0=200,T1=({open:s,onOpenChange:r,daemonInfo:c,onRefreshDaemon:u})=>{const[f,d]=h.useState(!1),[m,g]=h.useState(!1),[p,j]=h.useState(!1),[k,b]=h.useState(null),[v,z]=h.useState(!1),[E,$]=h.useState([]),[B,F]=h.useState(0),[Me,te]=h.useState(!1),[C,ie]=h.useState(!1),[de,W]=h.useState(!1),Ee=h.useRef(null),ne=h.useRef(!0),{toast:Ce}=X_(),pe=h.useCallback(async()=>{te(!0),ne.current=!0;try{const Ye=await C0(z0,0);$(Ye.lines||[]),F(Ye.total_lines||0)}catch(Ye){console.error("Failed to load daemon logs",Ye)}finally{te(!1)}},[]),V=()=>{const Ye=!de;W(Ye),Ye&&E.length===0&&pe()},Ie=h.useCallback(async()=>{if(C||E.length>=B)return;const Ye=Ee.current,Ne=Ye?Ye.scrollHeight:0,Xt=Ye?Ye.scrollTop:0;ie(!0);try{const xe=await C0(z0,E.length);xe.lines&&xe.lines.length>0&&($(We=>[...xe.lines,...We]),requestAnimationFrame(()=>{if(Ee.current){const We=Ee.current.scrollHeight;Ee.current.scrollTop=Xt+(We-Ne)}})),F(xe.total_lines||0)}catch(xe){console.error("Failed to load older daemon logs",xe)}finally{ie(!1)}},[C,E.length,B]),he=Ye=>{Ye.currentTarget.scrollTop<=20&&!C&&E.length<B&&!Me&&Ie()};h.useEffect(()=>{s?pe():W(!1)},[s,pe]),h.useEffect(()=>{E.length>0&&ne.current&&de&&(Ee.current&&(Ee.current.scrollTop=Ee.current.scrollHeight),ne.current=!1)},[E,de]),h.useEffect(()=>{de&&E.length>0&&requestAnimationFrame(()=>{Ee.current&&(Ee.current.scrollTop=Ee.current.scrollHeight)})},[de]);const ue=()=>{navigator.clipboard.writeText(E.join(`
`)),z(!0),setTimeout(()=>z(!1),2e3)},G=(Ye,Ne)=>{navigator.clipboard.writeText(Ne),b(Ye),setTimeout(()=>b(null),2e3)},we=async()=>{g(!0);try{await $b(),Ce({title:"Daemon Installed",description:"Native OS background service registered and scheduled for 15-minute quota sync.",variant:"success"}),await u(),await pe()}catch(Ye){const Ne=Ye instanceof Error?Ye.message:"Failed to install daemon";Ce({title:"Installation Failed",description:Ne,variant:"destructive"})}finally{g(!1)}},fe=async()=>{j(!0);try{await Ub(),Ce({title:"Daemon Removed",description:"Background service unregistered from the host OS service manager.",variant:"success"}),await u(),await pe()}catch(Ye){const Ne=Ye instanceof Error?Ye.message:"Failed to uninstall daemon";Ce({title:"Uninstall Failed",description:Ne,variant:"destructive"})}finally{j(!1)}},Oe=async()=>{d(!0);try{const Ye=await Hb();Ce({title:"Quota Sync Completed",description:Ye.message||"Triggered quota cache refresh across all configured profiles.",variant:"success"}),await u(),await pe()}catch(Ye){const Ne=Ye instanceof Error?Ye.message:"Failed to run daemon task";Ce({title:"Execution Failed",description:Ne,variant:"destructive"})}finally{d(!1)}},Se=Ye=>{if(!Ye)return"Never";try{const Ne=new Date(Ye);if(isNaN(Ne.getTime())||Ne.getTime()<=0)return"Never";const Xt=Date.now()-Ne.getTime(),xe=Math.floor(Xt/(1e3*60));if(xe<1)return"Just now";if(xe<60)return`${xe}m ago`;const We=Math.floor(xe/60);return We<24?`${We}h ago`:`${Math.floor(We/24)}d ago`}catch{return"Never"}},Qe=!!(c!=null&&c.installed),Ut=!!(c!=null&&c.active);return o.jsx(Dp,{open:s,onOpenChange:r,children:o.jsxs($_,{className:"max-w-2xl max-h-[90vh] overflow-y-auto bg-card border-border text-foreground p-6 shadow-2xl rounded-geist",children:[o.jsxs(U_,{className:"space-y-1.5 pb-3 border-b border-border pr-8",children:[o.jsxs("div",{className:"flex items-center gap-2.5 flex-wrap",children:[o.jsxs(Y_,{className:"text-base font-semibold text-foreground flex items-center gap-2",children:[o.jsx(b0,{className:"h-4 w-4 text-foreground"}),o.jsx("span",{children:"OS Background Daemon"})]}),o.jsx(Du,{variant:"outline",className:"text-xs font-mono bg-secondary/40 text-muted-foreground border-border px-2 py-0.5 rounded-geist",children:"15m Quota Pre-Warm"})]}),o.jsx(q_,{className:"text-xs text-muted-foreground leading-relaxed",children:"Native scheduled background service (launchd on macOS / systemd on Linux) for pre-warming quota and rate-limit caches without resident memory overhead."})]}),o.jsxs("div",{className:"rounded-geist border border-border bg-background p-3 text-xs space-y-2",children:[o.jsxs("div",{className:"flex items-center justify-between",children:[o.jsx("span",{className:"text-[11px] uppercase tracking-wider font-mono text-muted-foreground",children:"Operational Status"}),Ut?o.jsxs(Du,{variant:"outline",className:"bg-secondary/40 text-muted-foreground border-border text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist",children:[o.jsx("span",{className:"h-1.5 w-1.5 rounded-full bg-status-success-solid"}),o.jsx("span",{children:"Active (Every 15m)"})]}):Qe?o.jsxs(Du,{variant:"outline",className:"bg-secondary/40 text-muted-foreground border-border text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist",children:[o.jsx("span",{className:"h-1.5 w-1.5 rounded-full bg-status-warning-solid"}),o.jsx("span",{children:"Installed (Inactive)"})]}):o.jsxs(Du,{variant:"outline",className:"bg-secondary/40 text-muted-foreground border-border text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist",children:[o.jsx("span",{className:"h-1.5 w-1.5 rounded-full bg-muted-foreground/50"}),o.jsx("span",{children:"Not Installed"})]})]}),o.jsxs("div",{className:"grid grid-cols-2 gap-2 font-mono text-[11px] pt-1",children:[o.jsxs("div",{className:"p-2 rounded-geist bg-card border border-border",children:[o.jsx("span",{className:"text-muted-foreground/70 block text-xs uppercase",children:"Service Label"}),o.jsx("span",{className:"text-foreground font-medium truncate block",title:(c==null?void 0:c.label)||"com.aim-cli.daemon",children:(c==null?void 0:c.label)||"com.aim-cli.daemon"})]}),o.jsxs("div",{className:"p-2 rounded-geist bg-card border border-border",children:[o.jsx("span",{className:"text-muted-foreground/70 block text-xs uppercase",children:"Schedule Interval"}),o.jsxs("span",{className:"text-foreground font-medium block",children:["Every ",c!=null&&c.interval_sec?`${c.interval_sec/60}m`:"15m"]})]})]})]}),o.jsxs("div",{className:"rounded-geist border border-border bg-background p-3 text-xs space-y-2.5 font-mono",children:[o.jsxs("div",{className:"flex items-center justify-between text-muted-foreground",children:[o.jsxs("span",{className:"flex items-center gap-1.5 text-[11px]",children:[o.jsx(mb,{className:"h-3.5 w-3.5 text-muted-foreground/70"})," Last Sync"]}),o.jsx("span",{className:"text-foreground tabular-nums",children:Se(c==null?void 0:c.last_run)})]}),(c==null?void 0:c.last_run_message)&&o.jsx("div",{className:"p-2.5 rounded-geist bg-card border border-border text-[11px] text-muted-foreground leading-relaxed break-words",title:c.last_run_message,children:c.last_run_message}),(c==null?void 0:c.config_path)&&o.jsxs("div",{className:"flex items-center justify-between gap-3 text-muted-foreground pt-1.5 border-t border-border",children:[o.jsxs("span",{className:"flex items-center gap-1.5 text-[11px] truncate flex-1 min-w-0",title:c.config_path,children:[o.jsx(hb,{className:"h-3 w-3 shrink-0 text-muted-foreground/70"}),o.jsx("span",{className:"truncate",children:c.config_path})]}),o.jsxs("button",{type:"button",onClick:()=>G("config",c.config_path),className:"text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors shrink-0",children:[k==="config"?o.jsx(i_,{className:"h-3 w-3 text-status-success"}):o.jsx(c_,{className:"h-3 w-3"}),k==="config"?"Copied":"Copy"]})]}),(c==null?void 0:c.log_path)&&o.jsxs("div",{className:"flex items-center justify-between gap-3 text-muted-foreground",children:[o.jsxs("span",{className:"flex items-center gap-1.5 text-[11px] truncate flex-1 min-w-0",title:c.log_path,children:[o.jsx(gb,{className:"h-3 w-3 shrink-0 text-muted-foreground/70"}),o.jsx("span",{className:"truncate",children:c.log_path})]}),o.jsxs("button",{type:"button",onClick:()=>G("log",c.log_path),className:"text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors shrink-0",children:[k==="log"?o.jsx(i_,{className:"h-3 w-3 text-status-success"}):o.jsx(c_,{className:"h-3 w-3"}),k==="log"?"Copied":"Copy"]})]})]}),o.jsxs("div",{className:"rounded-geist border border-border bg-background p-3 text-xs space-y-2.5 transition-all",children:[o.jsxs("div",{className:"flex items-center justify-between",children:[o.jsxs("div",{className:"flex items-center gap-2 select-none",children:[o.jsx(kp,{className:"h-3.5 w-3.5 text-muted-foreground"}),o.jsx("span",{className:"text-[11px] uppercase tracking-wider font-mono text-muted-foreground font-semibold",children:"Execution Logs"}),B>0&&o.jsx("span",{className:"text-xs font-mono bg-secondary/70 text-muted-foreground border border-border px-1.5 py-0.5 rounded-geist leading-none inline-flex items-center",children:de&&E.length<B?`${E.length} of ${B} lines`:`${B} lines`})]}),o.jsxs("div",{className:"flex items-center gap-1.5",children:[de&&o.jsxs(o.Fragment,{children:[o.jsxs("button",{type:"button",onClick:pe,disabled:Me,title:"Refresh logs",className:"text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors",children:[o.jsx($r,{className:Pt("h-3 w-3",Me&&"animate-spin")}),"Refresh"]}),E.length>0&&o.jsxs("button",{type:"button",onClick:ue,title:"Copy logs",className:"text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors",children:[v?o.jsx(i_,{className:"h-3 w-3 text-status-success"}):o.jsx(c_,{className:"h-3 w-3"}),v?"Copied":"Copy"]})]}),o.jsx("button",{type:"button",onClick:V,className:"text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors",children:de?"Collapse":"View Logs"})]})]}),de&&o.jsx("div",{ref:Ee,onScroll:he,className:"bg-[#050505] border border-[#222222] rounded-geist p-2.5 font-mono text-[11px] max-h-44 overflow-y-auto space-y-1 select-text scrollbar-thin",children:Me&&E.length===0?o.jsxs("div",{className:"py-6 text-center text-muted-foreground/60 flex items-center justify-center gap-2",children:[o.jsx($r,{className:"h-3.5 w-3.5 animate-spin"}),o.jsx("span",{children:"Loading daemon logs..."})]}):E.length===0?o.jsx("div",{className:"py-6 text-center text-muted-foreground/60",children:"No daemon logs recorded yet. Trigger a sync below to generate execution records."}):o.jsxs(o.Fragment,{children:[C&&o.jsxs("div",{className:"py-2 text-center text-muted-foreground/60 flex items-center justify-center gap-1.5 border-b border-[#1c1c1c] mb-2 font-mono text-xs",children:[o.jsx($r,{className:"h-3 w-3 animate-spin text-primary"}),o.jsx("span",{children:"Loading earlier records..."})]}),E.length>=B&&B>0&&o.jsxs("div",{className:"py-2 text-center text-muted-foreground/40 select-none border-b border-[#1c1c1c] mb-2 font-mono text-xs",children:["— Beginning of execution logs (",B," lines) —"]}),E.map((Ye,Ne)=>{const{timestamp:Xt,level:xe,message:We}=E1(Ye);return o.jsxs("div",{className:"leading-relaxed break-all flex items-baseline gap-2 font-mono text-[11px] py-0.5",children:[Xt&&o.jsxs("span",{className:"text-[#666666] select-none shrink-0 font-mono text-xs",children:["[",Xt,"]"]}),o.jsx("span",{className:Pt("font-semibold select-none shrink-0 text-xs px-1.5 py-0.5 rounded-sm uppercase tracking-wide",xe==="ERROR"&&"text-red-400 bg-red-950/50 border border-red-800/50",xe==="WARN"&&"text-yellow-400 bg-yellow-950/50 border border-yellow-700/50",xe==="INFO"&&"text-sky-400 bg-sky-950/40 border border-sky-800/40"),children:xe}),o.jsx("span",{className:Pt("flex-1",xe==="ERROR"&&"text-red-300 font-medium",xe==="WARN"&&"text-yellow-300 font-medium",xe==="INFO"&&"text-[#ededed]"),children:We})]},Ne)})]})})]}),o.jsxs(H_,{className:"pt-2 flex items-center justify-between sm:justify-between w-full",children:[Qe?o.jsx($o,{type:"button",variant:"outline",size:"sm",disabled:p||f,onClick:fe,className:"border-border bg-card hover:bg-status-danger-bg hover:border-status-danger-border text-muted-foreground hover:text-status-danger text-xs font-sans rounded-geist",children:p?o.jsxs(o.Fragment,{children:[o.jsx($r,{className:"h-3.5 w-3.5 mr-1.5 animate-spin"}),"Uninstalling..."]}):o.jsxs(o.Fragment,{children:[o.jsx(pb,{className:"h-3.5 w-3.5 mr-1.5"}),"Uninstall Service"]})}):o.jsx($o,{type:"button",variant:"outline",size:"sm",onClick:()=>r(!1),className:"border-border bg-card hover:bg-secondary text-muted-foreground hover:text-foreground text-xs font-mono rounded-geist",children:"Close"}),o.jsx("div",{className:"flex items-center gap-2",children:Qe?o.jsx($o,{type:"button",size:"sm",disabled:f||p,onClick:Oe,className:"bg-primary hover:bg-primary/90 text-primary-foreground text-xs font-sans font-medium shadow-sm transition-all cursor-pointer rounded-geist",children:f?o.jsxs(o.Fragment,{children:[o.jsx($r,{className:"h-3.5 w-3.5 mr-1.5 animate-spin"}),"Syncing Quotas..."]}):o.jsxs(o.Fragment,{children:[o.jsx(yb,{className:"h-3.5 w-3.5 mr-1.5"}),"Sync Quotas Now"]})}):o.jsx($o,{type:"button",size:"sm",disabled:m,onClick:we,className:"bg-primary hover:bg-primary/90 text-primary-foreground text-xs font-sans font-medium shadow-sm transition-all cursor-pointer rounded-geist",children:m?o.jsxs(o.Fragment,{children:[o.jsx($r,{className:"h-3.5 w-3.5 mr-1.5 animate-spin"}),"Installing..."]}):o.jsxs(o.Fragment,{children:[o.jsx(b0,{className:"h-3.5 w-3.5 mr-1.5"}),"Install & Start Daemon"]})})})]})]})})},Pu=h.forwardRef(({className:s,type:r,...c},u)=>o.jsx("input",{type:r,className:Pt("flex h-9 w-full rounded-geist border border-border bg-[#0e0e0e] px-3 py-1.5 text-xs text-foreground placeholder:text-muted-foreground/60 transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3] focus-visible:border-[#0070f3] disabled:cursor-not-allowed disabled:opacity-50",s),ref:u,...c}));Pu.displayName="Input";const Bp=h.forwardRef(({className:s,checked:r=!1,onCheckedChange:c,disabled:u,onClick:f,size:d="default",...m},g)=>{const p=k=>{f==null||f(k),!u&&c&&c(!r)},j=d==="sm";return o.jsx("button",{type:"button",role:"switch","aria-checked":r,disabled:u,ref:g,onClick:p,className:Pt("peer inline-flex shrink-0 cursor-pointer items-center rounded-full border border-[#2e2e2e] transition-colors duration-200 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3] disabled:cursor-not-allowed disabled:opacity-50",j?"h-4 w-7":"h-5 w-9",r?"bg-[#0070f3] border-transparent":"bg-[#1c1c1c]",s),...m,children:o.jsx("span",{className:Pt("pointer-events-none block rounded-full transition-all duration-200 shadow-sm",j?"h-2.5 w-2.5":"h-3.5 w-3.5",r?j?"translate-x-[14px] bg-white":"translate-x-[18px] bg-white":"translate-x-[2px] bg-[#666666]")})})});Bp.displayName="Switch";function R1({open:s,onOpenChange:r,onProfileCreated:c}){const[u,f]=h.useState({agent:"claude",name:"",email:"",clone_from:""}),[d,m]=h.useState(!1),[g,p]=h.useState(!1),{toast:j}=X_();h.useEffect(()=>{s&&(f({agent:"claude",name:"",email:"",clone_from:""}),m(!1))},[s]);const k=async b=>{if(b.preventDefault(),!u.name.trim()){j({title:"Validation Error",description:"Profile name cannot be empty.",variant:"destructive"});return}p(!0);try{await Lb({...u,clone_from:d?u.clone_from:""}),j({title:"Profile Created",description:`Successfully added ${u.agent} profile "${u.name}".`,variant:"success"}),r(!1),c==null||c()}catch(v){const z=v instanceof Error?v.message:"Failed to create profile";j({title:"Creation Failed",description:z,variant:"destructive"})}finally{p(!1)}};return o.jsx(Dp,{open:s,onOpenChange:r,children:o.jsxs($_,{className:"border border-[#262626] bg-[#111111] text-[#ededed] rounded-geist sm:max-w-md shadow-2xl",children:[o.jsxs(U_,{className:"space-y-1.5 pb-2 border-b border-[#262626]",children:[o.jsx(Y_,{className:"text-base font-semibold tracking-tight text-[#ededed]",children:"Create New Agent Profile"}),o.jsx(q_,{className:"text-xs text-[#a1a1a1]",children:"Scaffold an isolated environment and configuration for your agent."})]}),o.jsxs("form",{onSubmit:k,className:"space-y-4 py-2",children:[o.jsxs("div",{className:"space-y-1.5",children:[o.jsx("label",{className:"text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]",children:"Target Agent"}),o.jsx("div",{className:"grid grid-cols-4 gap-2",children:["claude","codex","gemini","agy"].map(b=>o.jsx("button",{type:"button",onClick:()=>f({...u,agent:b}),className:`py-2 px-3 text-xs font-mono rounded-geist border text-center transition-all cursor-pointer ${u.agent===b?"border-[#ededed] bg-[#222222] text-[#ededed] font-bold shadow-sm":"border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#333333]"}`,children:b.toUpperCase()},b))})]}),o.jsxs("div",{className:"space-y-1.5",children:[o.jsx("label",{className:"text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]",children:"Profile Name *"}),o.jsx(Pu,{required:!0,placeholder:"e.g. work, sandbox, client-x",value:u.name,onChange:b=>f({...u,name:b.target.value}),className:"rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"})]}),o.jsxs("div",{className:"space-y-1.5",children:[o.jsx("label",{className:"text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]",children:"Associated Email (Optional)"}),o.jsx(Pu,{type:"email",placeholder:"dev@example.com",value:u.email,onChange:b=>f({...u,email:b.target.value}),className:"rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"})]}),o.jsxs("div",{className:"space-y-2 pt-2 border-t border-[#262626]",children:[o.jsxs("div",{className:"flex items-center justify-between py-0.5",children:[o.jsxs("div",{className:"space-y-0.5",children:[o.jsx("label",{htmlFor:"modal-clone-toggle",className:"text-xs font-medium text-[#ededed] cursor-pointer",children:"Clone from existing profile"}),o.jsx("p",{className:"text-[11px] text-[#a1a1a1]",children:"Copy authentication, credentials, and settings"})]}),o.jsx(Bp,{id:"modal-clone-toggle",checked:d,onCheckedChange:b=>{m(b),b||f(v=>({...v,clone_from:""}))}})]}),d&&o.jsxs("div",{className:"space-y-1.5 pt-1",children:[o.jsx("label",{className:"text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]",children:"Source Profile to Clone"}),o.jsx(Pu,{placeholder:"e.g. default",value:u.clone_from,onChange:b=>f({...u,clone_from:b.target.value}),className:"rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]",autoFocus:!0})]})]}),o.jsxs(H_,{className:"pt-3 border-t border-[#262626] flex gap-2 justify-end",children:[o.jsx($o,{type:"button",variant:"outline",onClick:()=>r(!1),disabled:g,className:"rounded-geist text-xs h-8 border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]",children:"Cancel"}),o.jsx($o,{type:"submit",disabled:g,className:"bg-[#ffffff] text-[#000000] hover:bg-[#eaeaea] rounded-geist text-xs h-8 font-medium shadow-sm transition-colors cursor-pointer",children:g?"Creating...":"Create Profile"})]})]})]})})}const z1=rb,$p=h.forwardRef(({className:s,...r},c)=>o.jsx(xp,{ref:c,className:Pt("fixed bottom-5 right-5 sm:bottom-6 sm:right-16 z-[100] flex max-h-screen w-full flex-col-reverse gap-2.5 md:max-w-[360px] pointer-events-none",s),...r}));$p.displayName=xp.displayName;const A1=I_("group pointer-events-auto relative flex w-full items-start justify-between gap-3 overflow-hidden rounded-geist border p-3.5 pr-8 shadow-[0_12px_36px_rgba(0,0,0,0.65)] backdrop-blur-md transition-all data-[swipe=cancel]:translate-x-0 data-[swipe=end]:translate-x-[var(--radix-toast-swipe-end-x)] data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)] data-[swipe=move]:transition-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[swipe=end]:animate-out data-[state=closed]:fade-out-80 data-[state=closed]:slide-out-to-right-full data-[state=open]:slide-in-from-top-full data-[state=open]:sm:slide-in-from-bottom-full",{variants:{variant:{default:"border-border/80 bg-[#121316]/95 text-foreground",destructive:"border-status-danger/40 bg-[#1a1214]/95 text-foreground",success:"border-status-success/35 bg-[#121815]/95 text-foreground",warning:"border-status-warning/40 bg-[#181612]/95 text-foreground"}},defaultVariants:{variant:"default"}}),Up=h.forwardRef(({className:s,variant:r,...c},u)=>o.jsx(bp,{ref:u,className:Pt(A1({variant:r}),s),...c}));Up.displayName=bp.displayName;const O1=h.forwardRef(({className:s,...r},c)=>o.jsx(vp,{ref:c,className:Pt("inline-flex h-8 shrink-0 items-center justify-center rounded-md border bg-transparent px-3 text-sm font-medium transition-colors hover:bg-secondary focus:outline-none focus:ring-1 focus:ring-ring disabled:pointer-events-none disabled:opacity-50 group-[.destructive]:border-muted/40 group-[.destructive]:hover:border-destructive/30 group-[.destructive]:hover:bg-destructive group-[.destructive]:hover:text-destructive-foreground group-[.destructive]:focus:ring-destructive",s),...r}));O1.displayName=vp.displayName;const Hp=h.forwardRef(({className:s,...r},c)=>o.jsx(wp,{ref:c,className:Pt("absolute right-2 top-2 rounded-md p-1 text-muted-foreground/60 transition-colors hover:text-foreground hover:bg-secondary/70 focus:opacity-100 focus:outline-none focus:ring-1 cursor-pointer",s),"toast-close":"",...r,children:o.jsx(Cp,{className:"h-3.5 w-3.5"})}));Hp.displayName=wp.displayName;const Yp=h.forwardRef(({className:s,...r},c)=>o.jsx(Sp,{ref:c,className:Pt("text-xs font-semibold text-foreground font-sans tracking-tight leading-snug",s),...r}));Yp.displayName=Sp.displayName;const qp=h.forwardRef(({className:s,...r},c)=>o.jsx(jp,{ref:c,className:Pt("text-[11px] text-muted-foreground font-sans leading-relaxed",s),...r}));qp.displayName=jp.displayName;function D1(){const{toasts:s}=X_(),r=c=>{switch(c){case"success":return o.jsx(vb,{className:"h-4 w-4 text-[#98c379] shrink-0 mt-0.5"});case"destructive":return o.jsx(bb,{className:"h-4 w-4 text-[#e06c75] shrink-0 mt-0.5"});case"warning":return o.jsx(Np,{className:"h-4 w-4 text-[#e5c07b] shrink-0 mt-0.5"});default:return o.jsx(xb,{className:"h-4 w-4 text-[#52a8ff] shrink-0 mt-0.5"})}};return o.jsxs(z1,{children:[s.map(function({id:c,title:u,description:f,action:d,variant:m,...g}){return o.jsxs(Up,{variant:m,...g,children:[o.jsxs("div",{className:"flex items-start gap-2.5 min-w-0 flex-1",children:[r(m),o.jsxs("div",{className:"grid gap-0.5 min-w-0 flex-1",children:[u&&o.jsx(Yp,{children:u}),f&&o.jsx(qp,{children:f})]})]}),d,o.jsx(Hp,{})]},c)}),o.jsx($p,{})]})}function Q_({size:s=16,variant:r="spoke",className:c,label:u}){const f=typeof s=="number"?s:s==="sm"?14:s==="lg"?24:16,d=r==="ring"?o.jsx("span",{style:{width:f,height:f},className:Pt("inline-block rounded-full border-2 border-white/15 border-t-white animate-spin shrink-0",c),role:"status","aria-label":u||"Loading"}):o.jsx("div",{className:Pt("geist-spinner inline-block relative text-[#ededed] shrink-0",c),style:{width:f,height:f},role:"status","aria-label":u||"Loading",children:Array.from({length:12}).map((m,g)=>o.jsx("div",{className:"geist-spinner-bar",style:{transform:`rotate(${g*30}deg) translate(0, -135%)`,animationDelay:`${-1.2+g*.1}s`}},g))});return u?o.jsxs("div",{className:"inline-flex items-center gap-2 text-xs font-sans text-muted-foreground/80 select-none",children:[d,o.jsx("span",{className:"truncate",children:u})]}):d}function A0({targetSessionId:s,onBack:r}){return o.jsxs("div",{className:"space-y-4 pb-12 animate-in fade-in duration-200",children:[o.jsxs("div",{className:"flex items-center justify-between gap-4",children:[o.jsxs($o,{variant:"ghost",size:"sm",onClick:r,className:"h-7 px-2.5 text-xs font-sans text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616] rounded-geist cursor-pointer",children:[o.jsx(wb,{className:"h-3.5 w-3.5 mr-1.5"}),"Back to Sessions"]}),o.jsxs("div",{className:"flex items-center gap-2 text-xs font-sans text-muted-foreground/80",children:[o.jsx(Q_,{size:14}),o.jsxs("span",{children:["Loading session ",s?`${s.slice(0,8)}…`:"…"]})]})]}),o.jsxs("div",{className:"rounded-geist border border-border bg-[#111111]/95 p-5 shadow-lg space-y-4",children:[o.jsxs("div",{className:"flex flex-col lg:flex-row lg:items-center justify-between gap-4",children:[o.jsxs("div",{className:"space-y-2 flex-1 min-w-0",children:[o.jsxs("div",{className:"flex items-center gap-2.5",children:[o.jsx("div",{className:"h-5 w-16 rounded-full bg-[#1c1c1c] animate-pulse"}),o.jsx("div",{className:"h-6 w-72 max-w-full rounded bg-[#1c1c1c] animate-pulse"})]}),o.jsx("div",{className:"h-4 w-4/5 max-w-xl rounded bg-[#161616] animate-pulse"})]}),o.jsxs("div",{className:"flex items-center gap-2 shrink-0",children:[o.jsx("div",{className:"h-8 w-24 rounded-geist bg-[#181818] border border-[#262626] animate-pulse"}),o.jsx("div",{className:"h-8 w-8 rounded-geist bg-[#181818] border border-[#262626] animate-pulse"})]})]}),o.jsxs("div",{className:"flex flex-wrap items-center gap-2 pt-2 border-t border-[#1c1c1c]",children:[o.jsx("div",{className:"h-6 w-24 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse"}),o.jsx("div",{className:"h-6 w-36 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse"}),o.jsx("div",{className:"h-6 w-44 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse"}),o.jsx("div",{className:"h-6 w-20 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse"})]})]}),o.jsxs("div",{className:"rounded-geist border border-border bg-[#0e0e0e] p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-5 w-40 rounded bg-[#181818] animate-pulse"}),o.jsx("div",{className:"h-6 w-48 rounded bg-[#161616] animate-pulse"})]}),o.jsx("div",{className:"h-7 w-64 max-w-full rounded bg-[#161616] border border-[#222222] animate-pulse"})]}),o.jsxs("div",{className:"space-y-3",children:[o.jsxs("div",{className:"rounded-geist border border-[#0070f3]/25 bg-[#0a0d14] p-4 space-y-3",children:[o.jsxs("div",{className:"flex items-center justify-between pb-2 border-b border-[#0070f3]/15",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-5 w-5 rounded-[4px] bg-[#0070f3]/20 animate-pulse"}),o.jsx("div",{className:"h-4 w-24 rounded bg-[#0070f3]/20 animate-pulse"}),o.jsx("div",{className:"h-3.5 w-12 rounded bg-[#161616] animate-pulse"})]}),o.jsx("div",{className:"h-4 w-16 rounded bg-[#161616] animate-pulse"})]}),o.jsxs("div",{className:"space-y-2",children:[o.jsx("div",{className:"h-4 w-5/6 rounded bg-[#161616] animate-pulse"}),o.jsx("div",{className:"h-4 w-2/3 rounded bg-[#161616] animate-pulse"})]})]}),o.jsxs("div",{className:"rounded-geist border border-[#1f1f1f] bg-[#111111] p-4 space-y-3",children:[o.jsxs("div",{className:"flex items-center justify-between pb-2 border-b border-[#1c1c1c]",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-5 w-5 rounded-[4px] bg-[#222222] animate-pulse"}),o.jsx("div",{className:"h-4 w-28 rounded bg-[#222222] animate-pulse"}),o.jsx("div",{className:"h-3.5 w-12 rounded bg-[#161616] animate-pulse"})]}),o.jsx("div",{className:"h-4 w-16 rounded bg-[#161616] animate-pulse"})]}),o.jsxs("div",{className:"space-y-2",children:[o.jsx("div",{className:"h-4 w-full rounded bg-[#161616] animate-pulse"}),o.jsx("div",{className:"h-4 w-11/12 rounded bg-[#161616] animate-pulse"}),o.jsx("div",{className:"h-4 w-3/4 rounded bg-[#161616] animate-pulse"})]})]})]})]})}function L1(){return o.jsx("div",{className:"flex items-center justify-center p-16 text-xs font-sans text-muted-foreground animate-in fade-in duration-200",children:o.jsxs("div",{className:"flex items-center gap-2.5",children:[o.jsx(Q_,{size:16}),o.jsx("span",{children:"Loading view…"})]})})}function B1({count:s=3}){return o.jsx("div",{className:"grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4",children:Array.from({length:s}).map((r,c)=>o.jsxs("div",{className:"rounded-geist border border-border bg-card p-4 space-y-4 shadow-xs",children:[o.jsxs("div",{className:"flex items-start justify-between gap-3",children:[o.jsxs("div",{className:"space-y-1.5 flex-1 min-w-0",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-4 w-4 rounded bg-muted animate-pulse shrink-0"}),o.jsx("div",{className:"h-4 w-28 rounded bg-muted animate-pulse"})]}),o.jsx("div",{className:"h-3 w-20 rounded bg-muted/60 animate-pulse"})]}),o.jsx("div",{className:"h-5 w-16 rounded-full bg-muted/60 animate-pulse shrink-0"})]}),o.jsxs("div",{className:"space-y-2 pt-1 border-t border-border/50",children:[o.jsx("div",{className:"h-3 w-24 rounded bg-muted/60 animate-pulse"}),o.jsxs("div",{className:"p-2.5 rounded-geist border border-border/60 bg-secondary/40 space-y-2",children:[o.jsxs("div",{className:"flex items-center justify-between gap-2",children:[o.jsxs("div",{className:"flex items-center gap-1.5",children:[o.jsx("div",{className:"h-4 w-12 rounded-full bg-muted animate-pulse"}),o.jsx("div",{className:"h-3.5 w-16 rounded bg-muted animate-pulse"})]}),o.jsx("div",{className:"h-5 w-16 rounded-geist bg-muted/80 animate-pulse"})]}),o.jsx("div",{className:"h-1.5 w-full rounded-full bg-muted/40 animate-pulse"})]})]}),o.jsxs("div",{className:"flex items-center gap-2 pt-2 border-t border-border/40",children:[o.jsx("div",{className:"h-3.5 w-14 rounded bg-muted/40 animate-pulse"}),o.jsx("div",{className:"h-3.5 w-12 rounded bg-muted/40 animate-pulse"}),o.jsx("div",{className:"h-3.5 w-16 rounded bg-muted/40 animate-pulse"})]})]},c))})}function $1(){return o.jsxs("div",{className:"space-y-6",children:[o.jsxs("div",{className:"flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border",children:[o.jsx("div",{className:"space-y-1",children:o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-4 w-4 rounded-full bg-muted animate-pulse"}),o.jsx("div",{className:"h-6 w-48 rounded bg-muted animate-pulse"}),o.jsx("div",{className:"h-5 w-14 rounded-full bg-muted/60 animate-pulse"})]})}),o.jsx("div",{className:"flex items-center gap-2 shrink-0",children:o.jsx("div",{className:"h-8 w-72 rounded-geist bg-secondary border border-border animate-pulse shrink-0"})})]}),o.jsxs("div",{className:"p-4 rounded-geist border border-border bg-card flex flex-col sm:flex-row sm:items-center justify-between gap-4",children:[o.jsxs("div",{className:"flex items-center gap-3 min-w-0",children:[o.jsx("div",{className:"h-8 w-8 rounded-full border border-border bg-secondary shrink-0 animate-pulse"}),o.jsxs("div",{className:"space-y-1.5",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-3.5 w-36 rounded bg-muted animate-pulse"}),o.jsx("div",{className:"h-2 w-2 rounded-full bg-muted/60 animate-pulse"})]}),o.jsx("div",{className:"h-3 w-64 max-w-full rounded bg-muted/50 animate-pulse"})]})]}),o.jsx("div",{className:"h-8 w-20 rounded-geist bg-secondary border border-border animate-pulse shrink-0"})]}),o.jsxs("div",{className:"flex items-center justify-between gap-3 pb-1",children:[o.jsx("div",{className:"h-9 w-96 rounded-geist bg-secondary/40 border border-border animate-pulse"}),o.jsx("div",{className:"hidden md:flex items-center gap-3",children:o.jsx("div",{className:"h-4 w-24 rounded bg-muted/40 animate-pulse"})})]}),o.jsxs("div",{className:"space-y-3",children:[o.jsxs("div",{className:"flex flex-col sm:flex-row sm:items-center justify-between gap-4",children:[o.jsxs("div",{className:"space-y-1",children:[o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-4 w-4 rounded bg-muted animate-pulse"}),o.jsx("div",{className:"h-4 w-36 rounded bg-muted animate-pulse"}),o.jsx("div",{className:"h-4 w-12 rounded-full bg-muted/60 animate-pulse"})]}),o.jsx("div",{className:"h-3 w-60 rounded bg-muted/40 animate-pulse"})]}),o.jsxs("div",{className:"flex items-center gap-2",children:[o.jsx("div",{className:"h-8 w-48 rounded-geist bg-secondary border border-border animate-pulse shrink-0"}),o.jsx("div",{className:"h-8 w-8 rounded-geist bg-secondary border border-border animate-pulse shrink-0"}),o.jsx("div",{className:"h-8 w-8 rounded-geist bg-secondary border border-border animate-pulse shrink-0"})]})]}),o.jsx(B1,{count:3})]})]})}function U1({count:s=3}){return o.jsx("div",{className:"grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 animate-in fade-in duration-200",children:Array.from({length:s}).map((r,c)=>o.jsxs("div",{className:"rounded-geist border border-border bg-card p-5 space-y-4 shadow-xs",children:[o.jsxs("div",{className:"flex items-start justify-between gap-3",children:[o.jsxs("div",{className:"flex items-center gap-2.5 min-w-0",children:[o.jsx("div",{className:"h-8 w-8 rounded-full bg-muted animate-pulse shrink-0"}),o.jsxs("div",{className:"space-y-1",children:[o.jsx("div",{className:"h-4 w-28 rounded bg-muted animate-pulse"}),o.jsx("div",{className:"h-3 w-20 rounded bg-muted/60 animate-pulse"})]})]}),o.jsx("div",{className:"h-5 w-16 rounded-full bg-muted/60 animate-pulse shrink-0"})]}),o.jsxs("div",{className:"p-3 rounded-geist bg-secondary/50 border border-border/60 space-y-2",children:[o.jsxs("div",{className:"flex items-center justify-between",children:[o.jsx("div",{className:"h-3 w-20 rounded bg-muted/80 animate-pulse"}),o.jsx("div",{className:"h-3 w-10 rounded bg-muted/80 animate-pulse"})]}),o.jsx("div",{className:"h-2 w-full rounded-full bg-muted/40 animate-pulse"})]}),o.jsxs("div",{className:"flex items-center justify-between pt-2 border-t border-border/40",children:[o.jsx("div",{className:"h-3 w-24 rounded bg-muted/40 animate-pulse"}),o.jsx("div",{className:"h-6 w-16 rounded-geist bg-muted/60 animate-pulse"})]})]},c))})}function H1({count:s=3}){return o.jsx("div",{className:"grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 animate-in fade-in duration-200",children:Array.from({length:s}).map((r,c)=>o.jsxs("div",{className:"rounded-geist border border-border bg-card p-5 space-y-3.5 shadow-xs",children:[o.jsxs("div",{className:"flex items-start justify-between gap-3",children:[o.jsxs("div",{className:"flex items-center gap-2.5 min-w-0",children:[o.jsx("div",{className:"h-7 w-7 rounded-geist bg-muted animate-pulse shrink-0"}),o.jsxs("div",{className:"space-y-1",children:[o.jsx("div",{className:"h-4 w-28 rounded bg-muted animate-pulse"}),o.jsx("div",{className:"h-3 w-20 rounded bg-muted/60 animate-pulse"})]})]}),o.jsx("div",{className:"h-5 w-14 rounded-full bg-muted/60 animate-pulse shrink-0"})]}),o.jsx("div",{className:"h-3 w-3/4 rounded bg-muted/40 animate-pulse"}),o.jsxs("div",{className:"flex items-center gap-2 pt-2 border-t border-border/40",children:[o.jsx("div",{className:"h-5 w-16 rounded-[4px] bg-secondary border border-border animate-pulse"}),o.jsx("div",{className:"h-5 w-20 rounded-[4px] bg-secondary border border-border animate-pulse"})]})]},c))})}function Y1({profileName:s}){return o.jsxs("div",{className:"space-y-6 animate-in fade-in duration-200",children:[o.jsxs("div",{className:"flex flex-col sm:flex-row sm:items-center justify-between gap-4",children:[o.jsxs("div",{className:"flex items-center gap-3",children:[o.jsx("div",{className:"h-7 w-32 rounded-geist bg-[#181818] animate-pulse"}),o.jsx("div",{className:"h-5 w-16 rounded-full bg-[#1c1c1c] animate-pulse"}),o.jsx("div",{className:"h-5 w-24 rounded-full bg-[#1c1c1c] animate-pulse"})]}),o.jsxs("div",{className:"flex items-center gap-2 text-xs font-sans text-muted-foreground/80",children:[o.jsx(Q_,{size:14}),o.jsxs("span",{children:["Loading profile ",s?`"${s}"`:"","…"]})]})]}),o.jsx("div",{className:"h-9 w-80 rounded-geist bg-[#141414] border border-[#222222] animate-pulse"}),o.jsxs("div",{className:"grid grid-cols-1 md:grid-cols-2 gap-4",children:[o.jsxs("div",{className:"rounded-geist border border-[#2a2a2a] bg-[#0c0c0c] p-5 space-y-4",children:[o.jsx("div",{className:"h-5 w-32 rounded bg-[#181818] animate-pulse"}),o.jsx("div",{className:"h-10 w-full rounded bg-[#141414] animate-pulse"}),o.jsx("div",{className:"h-4 w-2/3 rounded bg-[#161616] animate-pulse"})]}),o.jsxs("div",{className:"rounded-geist border border-[#2a2a2a] bg-[#0c0c0c] p-5 space-y-4",children:[o.jsx("div",{className:"h-5 w-32 rounded bg-[#181818] animate-pulse"}),o.jsx("div",{className:"h-10 w-full rounded bg-[#141414] animate-pulse"}),o.jsx("div",{className:"h-4 w-2/3 rounded bg-[#161616] animate-pulse"})]})]})]})}class q1 extends h.Component{constructor(c){super(c);y0(this,"handleRetry",()=>{this.setState({hasError:!1,error:null}),this.props.onReset&&this.props.onReset()});this.state={hasError:!1,error:null}}static getDerivedStateFromError(c){return{hasError:!0,error:c}}componentDidCatch(c,u){console.error("ErrorBoundary caught an error:",c,u)}render(){var c;return this.state.hasError?o.jsxs("div",{className:"flex flex-col items-center justify-center min-h-[360px] p-8 text-center rounded-geist border border-status-danger-border/60 bg-status-danger-bg/20 space-y-4",children:[o.jsx("div",{className:"h-12 w-12 rounded-full bg-status-danger-bg border border-status-danger-border flex items-center justify-center text-status-danger",children:o.jsx(Np,{className:"h-6 w-6"})}),o.jsxs("div",{className:"space-y-1.5 max-w-md",children:[o.jsx("h3",{className:"text-base font-semibold text-foreground font-sans",children:this.props.fallbackTitle||"Something went wrong in this view"}),o.jsx("p",{className:"text-xs text-muted-foreground font-mono leading-relaxed break-words",children:((c=this.state.error)==null?void 0:c.message)||"An unexpected rendering error occurred."})]}),o.jsxs("div",{className:"flex items-center gap-2 pt-2",children:[o.jsxs($o,{size:"sm",variant:"outline",onClick:this.handleRetry,className:"h-8 text-xs font-sans rounded-geist border-border bg-secondary hover:bg-muted text-foreground cursor-pointer flex items-center gap-1.5",children:[o.jsx(Sb,{className:"h-3.5 w-3.5"}),o.jsx("span",{children:"Try Again"})]}),o.jsx($o,{size:"sm",onClick:()=>window.location.reload(),className:"h-8 text-xs font-sans rounded-geist bg-foreground text-background hover:bg-foreground/90 font-medium cursor-pointer",children:"Reload Page"})]})]}):this.props.children}}var Ip=["data-feedback-toolbar","data-annotation-popup","data-annotation-marker"],p_=Ip.flatMap(s=>[`:not([${s}])`,`:not([${s}] *)`]).join(""),E_="feedback-freeze-styles",T_="__agentation_freeze";function I1(){return typeof window>"u"?{frozen:!1,installed:!0,origSetTimeout:setTimeout,origSetInterval:setInterval,origRAF:r=>0,pausedAnimations:[],frozenTimeoutQueue:[],frozenRAFQueue:[]}:window[T_]??{frozen:!1,installed:!1,origSetTimeout:window.setTimeout.bind(window),origSetInterval:window.setInterval.bind(window),origRAF:window.requestAnimationFrame.bind(window),pausedAnimations:[],frozenTimeoutQueue:[],frozenRAFQueue:[]}}var Tt=I1();function Xp(){if(typeof window>"u")return;const s=window;Tt=s[T_]??(s[T_]=Tt),!Tt.installed&&(window.setTimeout=(r,c,...u)=>typeof r=="string"?Tt.origSetTimeout(r,c):Tt.origSetTimeout((...f)=>{Tt.frozen?Tt.frozenTimeoutQueue.push(()=>r(...f)):r(...f)},c,...u),window.setInterval=(r,c,...u)=>typeof r=="string"?Tt.origSetInterval(r,c):Tt.origSetInterval((...f)=>{Tt.frozen||r(...f)},c,...u),window.requestAnimationFrame=r=>Tt.origRAF(c=>{Tt.frozen?Tt.frozenRAFQueue.push(r):r(c)}),Tt.installed=!0)}var Ht=Tt.origSetTimeout,Qp=Tt.origSetInterval,nd=Tt.origRAF;function X1(s){return s?Ip.some(r=>{var c;return!!((c=s.closest)!=null&&c.call(s,`[${r}]`))}):!1}function Q1(){if(typeof document>"u"||(Xp(),Tt.frozen))return;Tt.frozen=!0,Tt.frozenTimeoutQueue=[],Tt.frozenRAFQueue=[];let s=document.getElementById(E_);s||(s=document.createElement("style"),s.id=E_),s.textContent=`
    *${p_},
    *${p_}::before,
    *${p_}::after {
      animation-play-state: paused !important;
      transition: none !important;
    }
  `,document.head.appendChild(s),Tt.pausedAnimations=[];try{document.getAnimations().forEach(r=>{var u;if(r.playState!=="running")return;const c=(u=r.effect)==null?void 0:u.target;X1(c)||(r.pause(),Tt.pausedAnimations.push(r))})}catch{}document.querySelectorAll("video").forEach(r=>{r.paused||(r.dataset.wasPaused="false",r.pause())})}function O0(){var c;if(typeof document>"u"||!Tt.frozen)return;Tt.frozen=!1;const s=Tt.frozenTimeoutQueue;Tt.frozenTimeoutQueue=[];for(const u of s)Tt.origSetTimeout(()=>{if(Tt.frozen){Tt.frozenTimeoutQueue.push(u);return}try{u()}catch(f){console.warn("[agentation] Error replaying queued timeout:",f)}},0);const r=Tt.frozenRAFQueue;Tt.frozenRAFQueue=[];for(const u of r)Tt.origRAF(f=>{if(Tt.frozen){Tt.frozenRAFQueue.push(u);return}u(f)});for(const u of Tt.pausedAnimations)try{u.play()}catch(f){console.warn("[agentation] Error resuming animation:",f)}Tt.pausedAnimations=[],(c=document.getElementById(E_))==null||c.remove(),document.querySelectorAll("video").forEach(u=>{u.dataset.wasPaused==="false"&&(u.play().catch(()=>{}),delete u.dataset.wasPaused)})}function D0(){const s=h.useMemo(()=>{let r=0;const c=new Set,u=()=>(c.forEach(clearTimeout),c.clear(),++r),f=m=>m===r;return{start:u,isCurrent:f,schedule:(m,g,p)=>{if(!f(m))return;const j=Ht(()=>{c.delete(j),f(m)&&g()},p);c.add(j)}}},[]);return h.useEffect(()=>()=>{s.start()},[s]),s}function G1(s,r,c,u){const f=j=>u.get(j.id)??j.id,d=new Map(s.map(j=>[f(j),j])),m=new Map(r.map(j=>[f(j),j])),g=new Set(c.map(j=>j.id)),p=[];for(const j of c){const k=d.get(j.id),b=m.get(j.id);if(k&&!b)continue;const v=b&&k&&b.comment!==k.comment;p.push(v?{...j,comment:b.comment}:j)}for(const[j,k]of m)!d.has(j)&&!g.has(j)&&p.push(k);return p}function mc(s){if(s.tagName!=="IFRAME")return null;try{return s.contentDocument}catch{return null}}function Xl(s,r=document){var c;if(s===r)return null;try{return(c=s.defaultView)==null?void 0:c.frameElement}catch{return null}}function G_(s){return s.nodeType===11&&"host"in s}function Qr(s){var k;const r=s.getBoundingClientRect(),c=s.offsetWidth?r.width/s.offsetWidth:1,u=s.offsetHeight?r.height/s.offsetHeight:1,f=(k=s.ownerDocument.defaultView)==null?void 0:k.getComputedStyle(s),d=b=>parseFloat(b||"0")||0,m=d(f==null?void 0:f.paddingLeft),g=d(f==null?void 0:f.paddingTop),p=s.clientWidth-m-d(f==null?void 0:f.paddingRight),j=s.clientHeight-g-d(f==null?void 0:f.paddingBottom);return{x:r.left+(s.clientLeft+m)*c,y:r.top+(s.clientTop+g)*u,sx:c,sy:u,width:p*c,height:j*u}}function L0(s){try{const r=new URL(s);return r.origin+r.pathname}catch{return s}}function V1(s,r,c,u=document){let f=Xl(s,u);for(;f;){const d=Qr(f);r=d.x+r*d.sx,c=d.y+c*d.sy,f=Xl(f.ownerDocument,u)}return{x:r,y:c}}function Un(s,r=document){const c=s.getBoundingClientRect();return Gp(s.ownerDocument,c,r)}function Gp(s,r,c){if(!Xl(s,c))return r;let u=r.left,f=r.top,d=r.right,m=r.bottom,g=Xl(s,c);for(;g;){const p=Qr(g);u=Math.max(p.x,p.x+u*p.sx),f=Math.max(p.y,p.y+f*p.sy),d=Math.min(p.x+p.width,p.x+d*p.sx),m=Math.min(p.y+p.height,p.y+m*p.sy),g=Xl(g.ownerDocument,c)}return new DOMRect(u,f,Math.max(0,d-u),Math.max(0,m-f))}function od(s){const r=[];for(const c of s.querySelectorAll("*"))c.tagName==="IFRAME"&&r.push(c),c.shadowRoot&&c.tagName!=="AGENTATION-TOOLBAR"&&r.push(...od(c.shadowRoot));return r}function F1(s,r,c,u=document){const f=[];for(let b=Xl(s.ownerDocument,u);b;b=Xl(b.ownerDocument,u))f.unshift(b);if(!f.length)return;const d=f.map(b=>{var z;const v=Qr(b);return r=(r-v.x)/v.sx,c=(c-v.y)/v.sy,{index:od(b.ownerDocument).indexOf(b),id:b.id||void 0,url:((z=mc(b))==null?void 0:z.URL)??""}}),m=s.ownerDocument.defaultView;let g=!1;for(let b=s;b;b=b.parentElement)if(["fixed","sticky"].includes(m.getComputedStyle(b).position)){g=!0;break}const p=g?0:m.scrollX,j=g?0:m.scrollY,k=s.getBoundingClientRect();return{path:d,x:r+p,y:c+j,fixed:g,boundingBox:{x:k.left+p,y:k.top+j,width:k.width,height:k.height}}}function W1(s=document){const r=new Map,c=u=>{let f=r.get(u);return f||(f=od(u),r.set(u,f)),f};return u=>K1(u,s,c)}function K1(s,r=document,c=od){const u=s.frame;if(!u)return s;let f=r;for(const z of u.path){const E=c(f),$=z.id?E.find(F=>F.id===z.id):E[z.index],B=$&&mc($);if(!B||L0(B.URL)!==L0(z.url))return null;f=B}const d=f.defaultView,m=u.fixed?0:d.scrollX,g=u.fixed?0:d.scrollY;let p=u.x-m,j=u.y-g;for(let z=f,E=Xl(f,r);E;E=Xl(z,r)){const $=z.defaultView;if(p<0||j<0||p>$.innerWidth||j>$.innerHeight)return null;const B=Qr(E);if(B.width<=0||B.height<=0)return null;p=B.x+p*B.sx,j=B.y+j*B.sy,z=E.ownerDocument}const k=u.boundingBox,b=Gp(f,new DOMRect(k.x-m,k.y-g,k.width,k.height),r),v=r.defaultView;return{...s,x:p/v.innerWidth*100,y:j+(s.isFixed?0:v.scrollY),boundingBox:{x:b.x,y:b.y+(s.isFixed?0:v.scrollY),width:b.width,height:b.height}}}function Z1(s,r,c){if(!("clientX"in s)||r===c)return s;const u=V1(r,s.clientX,s.clientY,c);return new Proxy(s,{get(f,d){if(d==="clientX")return u.x;if(d==="clientY")return u.y;const m=Reflect.get(f,d,f);return typeof m=="function"?m.bind(f):m}})}function J1(s,r){const c=new Set([s]),u=new Set;let f=[],d=new Set,m=!1,g=!1,p;const j=(b,v)=>{const z=E=>b.listener(Z1(E,v,s));b.handlers.set(v,z),v.addEventListener(b.type,z,b.options)},k=()=>{if(m=!1,!g)return;const b=new Set,v=new Set,z=[],E=$=>{z.push($);for(const B of $.querySelectorAll("*"))if(!B.matches("agentation-toolbar, [data-agentation-portal]")&&(B.shadowRoot&&E(B.shadowRoot),B.tagName==="IFRAME")){v.add(B);const F=mc(B);F&&!b.has(F)&&(b.add(F),E(F))}};b.add(s),E(s);for(const $ of c)if(!b.has($)){for(const B of u){const F=B.handlers.get($);F&&$.removeEventListener(B.type,F,B.options),B.handlers.delete($)}c.delete($)}for(const $ of b)if(!c.has($)){c.add($);for(const B of u)j(B,$)}for(const $ of d)v.has($)||$.removeEventListener("load",k);for(const $ of v)d.has($)||$.addEventListener("load",k);d=v,p==null||p.disconnect(),d.forEach($=>p==null?void 0:p.observe($)),f.forEach($=>$.disconnect()),f=z.map($=>{const B=new MutationObserver(F=>{F.some(te=>[...te.addedNodes,...te.removedNodes].some(C=>C.nodeType===1&&!C.closest("agentation-toolbar, [data-agentation-portal]")&&(C.tagName==="IFRAME"||!!C.shadowRoot||!!C.querySelector("iframe"))))&&!m&&(m=!0,queueMicrotask(k))});return B.observe($,{childList:!0,subtree:!0}),B}),r==null||r()};return{start(){g||(g=!0,typeof ResizeObserver=="function"&&(p=new ResizeObserver(()=>r==null?void 0:r())),k())},stop(){g=!1,p==null||p.disconnect(),p=void 0,f.forEach(b=>b.disconnect()),f=[];for(const b of d)b.removeEventListener("load",k);d.clear();for(const b of u)for(const[v,z]of b.handlers)v.removeEventListener(b.type,z,b.options);u.clear(),c.clear(),c.add(s)},addEventListener(b,v,z){const E={type:b,listener:v,options:z,handlers:new Map};u.add(E);for(const $ of c)j(E,$)},removeEventListener(b,v,z){for(const E of u)if(E.type===b&&E.listener===v){for(const[$,B]of E.handlers)$.removeEventListener(b,B,E.options);u.delete(E)}},querySelectorAll(b){return[...c].flatMap(v=>[...v.querySelectorAll(b)])}}}var V_=["data-testid","data-test","data-qa","data-cy","data-component"];function ic(s,r=V_){const c={};for(const u of[...new Set(r)].slice(0,16)){if(!/^[a-zA-Z_][\w:.-]*$/.test(u))continue;const f=s.getAttribute(u);f!=null&&f.length<=500&&Object.defineProperty(c,u,{value:f,enumerable:!0})}return c}function P1(s,r=V_){return Object.entries(ic(s,r)).filter(([c,u])=>/^data-[a-z0-9_-]+$/.test(c)&&u.length<=120).slice(0,2).map(([c,u])=>`[${c}="${u.replace(/[\\"\n\r\f\0]/g,f=>f==="\\"||f==='"'?`\\${f}`:`\\${f.charCodeAt(0).toString(16)} `)}"]`).join("")}function Ir(s){if(s.parentElement)return s.parentElement;const r=s.getRootNode();return G_(r)?r.host:null}function ol(s,r){let c=s;for(;c;){if(c.matches(r))return c;c=Ir(c)}return null}function Vp(s,r=4,c){const u=[];let f=s,d=0;for(;f&&d<r;){const g=f.tagName.toLowerCase();if(g==="html"||g==="body"){u.length===0&&u.push(g);break}let p=g;if(f.id)p=`#${f.id}`;else if(f.className&&typeof f.className=="string"){const k=f.className.split(/\s+/).find(b=>b.length>2&&!b.match(/^[a-z]{1,2}$/)&&!b.match(/[A-Z0-9]{5,}/));k&&(p=`.${k.split("_")[0]}`)}p+=P1(f,c);const j=Ir(f);!f.parentElement&&j&&(p=`⟨shadow⟩ ${p}`),u.unshift(p),f=j,d++}const m=Xl(s.ownerDocument);return(m?Vp(m,2)+" > ⟨iframe⟩ ":"")+u.join(" > ")}function ev(s){var c;let r="";for(const u of s.childNodes)if(u.nodeType===Node.TEXT_NODE){const f=(c=u.textContent)==null?void 0:c.trim();f&&(r+=(r?" ":"")+f)}return r}function qr(s,r){var f,d,m,g,p,j,k,b;const c=Vp(s,4,r);if(s.dataset.element)return{name:s.dataset.element,path:c};const u=s.tagName.toLowerCase();if(["path","circle","rect","line","g"].includes(u)){const v=ol(s,"svg");if(v){const z=Ir(v);if((z==null?void 0:z.namespaceURI)==="http://www.w3.org/1999/xhtml")return{name:`graphic in ${qr(z).name}`,path:c}}return{name:"graphic element",path:c}}if(u==="svg"){const v=Ir(s);if((v==null?void 0:v.tagName.toLowerCase())==="button"){const z=(f=v.textContent)==null?void 0:f.trim();return{name:z?`icon in "${z}" button`:"button icon",path:c}}return{name:"icon",path:c}}if(u==="button"){const v=(d=s.textContent)==null?void 0:d.trim(),z=s.getAttribute("aria-label");return z?{name:`button [${z}]`,path:c}:{name:v?`button "${v.slice(0,25)}"`:"button",path:c}}if(u==="a"){const v=(m=s.textContent)==null?void 0:m.trim(),z=s.getAttribute("href");return v?{name:`link "${v.slice(0,25)}"`,path:c}:z?{name:`link to ${z.slice(0,30)}`,path:c}:{name:"link",path:c}}if(u==="input"){const v=s.getAttribute("type")||"text",z=s.getAttribute("placeholder"),E=s.getAttribute("name");return z?{name:`input "${z}"`,path:c}:E?{name:`input [${E}]`,path:c}:{name:`${v} input`,path:c}}if(["h1","h2","h3","h4","h5","h6"].includes(u)){const v=(g=s.textContent)==null?void 0:g.trim();return{name:v?`${u} "${v.slice(0,35)}"`:u,path:c}}if(u==="p"){const v=(p=s.textContent)==null?void 0:p.trim();return v?{name:`paragraph: "${v.slice(0,40)}${v.length>40?"...":""}"`,path:c}:{name:"paragraph",path:c}}if(u==="span"||u==="label"){const v=(j=s.textContent)==null?void 0:j.trim();return v&&v.length<40?{name:`"${v}"`,path:c}:{name:u,path:c}}if(u==="li"){const v=(k=s.textContent)==null?void 0:k.trim();return v&&v.length<40?{name:`list item: "${v.slice(0,35)}"`,path:c}:{name:"list item",path:c}}if(u==="blockquote")return{name:"blockquote",path:c};if(u==="code"){const v=(b=s.textContent)==null?void 0:b.trim();return v&&v.length<30?{name:`code: \`${v}\``,path:c}:{name:"code",path:c}}if(u==="pre")return{name:"code block",path:c};if(u==="img"){const v=s.getAttribute("alt");return{name:v?`image "${v.slice(0,30)}"`:"image",path:c}}if(u==="video")return{name:"video",path:c};if(["div","section","article","nav","header","footer","aside","main"].includes(u)){const v=s.className,z=s.getAttribute("role"),E=s.getAttribute("aria-label");if(E)return{name:`${u} [${E}]`,path:c};if(z)return{name:`${z}`,path:c};const $=ev(s);if($&&$.length<50)return{name:`"${$}"`,path:c};if(typeof v=="string"&&v){const B=v.split(/[\s_-]+/).map(F=>F.replace(/[A-Z0-9]{5,}.*$/,"")).filter(F=>F.length>2&&!/^[a-z]{1,2}$/.test(F)).slice(0,2);if(B.length>0)return{name:B.join(" "),path:c}}return{name:u==="div"?"container":u,path:c}}return{name:u,path:c}}function nc(s){var d,m,g;const r=[],c=(d=s.textContent)==null?void 0:d.trim();c&&c.length<100&&r.push(c);const u=s.previousElementSibling;if(u){const p=(m=u.textContent)==null?void 0:m.trim();p&&p.length<50&&r.unshift(`[before: "${p.slice(0,40)}"]`)}const f=s.nextElementSibling;if(f){const p=(g=f.textContent)==null?void 0:g.trim();p&&p.length<50&&r.push(`[after: "${p.slice(0,40)}"]`)}return r.join(" ")}function Lu(s){const r=Ir(s);if(!r)return"";const c=s.getRootNode(),f=(G_(c)&&s.parentElement?Array.from(s.parentElement.children):Array.from(r.children)).filter(k=>k!==s&&k.namespaceURI==="http://www.w3.org/1999/xhtml");if(f.length===0)return"";const d=f.slice(0,4).map(k=>{var E;const b=k.tagName.toLowerCase(),v=k.className;let z="";if(typeof v=="string"&&v){const $=v.split(/\s+/).map(B=>B.replace(/[_][a-zA-Z0-9]{5,}.*$/,"")).find(B=>B.length>2&&!/^[a-z]{1,2}$/.test(B));$&&(z=`.${$}`)}if(b==="button"||b==="a"){const $=(E=k.textContent)==null?void 0:E.trim().slice(0,15);if($)return`${b}${z} "${$}"`}return`${b}${z}`});let g=r.tagName.toLowerCase();if(typeof r.className=="string"&&r.className){const k=r.className.split(/\s+/).map(b=>b.replace(/[_][a-zA-Z0-9]{5,}.*$/,"")).find(b=>b.length>2&&!/^[a-z]{1,2}$/.test(b));k&&(g=`.${k}`)}const p=r.children.length,j=p>d.length+1?` (${p} total in ${g})`:"";return d.join(", ")+j}function lc(s){const r=s.className;return typeof r!="string"||!r?"":r.split(/\s+/).filter(u=>u.length>0).map(u=>{const f=u.match(/^([a-zA-Z][a-zA-Z0-9_-]*?)(?:_[a-zA-Z0-9]{5,})?$/);return f?f[1]:u}).filter((u,f,d)=>d.indexOf(u)===f).join(", ")}var Fp=new Set(["none","normal","auto","0px","rgba(0, 0, 0, 0)","transparent","static","visible"]),tv=new Set(["p","span","h1","h2","h3","h4","h5","h6","label","li","td","th","blockquote","figcaption","caption","legend","dt","dd","pre","code","em","strong","b","i","a","time","cite","q"]),nv=new Set(["input","textarea","select"]),lv=new Set(["img","video","canvas","svg"]),ov=new Set(["div","section","article","nav","header","footer","aside","main","ul","ol","form","fieldset"]);function Bu(s){if(typeof window>"u")return{};const r=(s.ownerDocument.defaultView??window).getComputedStyle(s),c={},u=s.tagName.toLowerCase();let f;tv.has(u)?f=["color","fontSize","fontWeight","fontFamily","lineHeight"]:u==="button"||u==="a"&&s.getAttribute("role")==="button"?f=["backgroundColor","color","padding","borderRadius","fontSize"]:nv.has(u)?f=["backgroundColor","color","padding","borderRadius","fontSize"]:lv.has(u)?f=["width","height","objectFit","borderRadius"]:ov.has(u)?f=["display","padding","margin","gap","backgroundColor"]:f=["color","fontSize","margin","padding","backgroundColor"];for(const d of f){const m=d.replace(/([A-Z])/g,"-$1").toLowerCase(),g=r.getPropertyValue(m);g&&!Fp.has(g)&&(c[d]=g)}return c}var sv=["color","backgroundColor","borderColor","fontSize","fontWeight","fontFamily","lineHeight","letterSpacing","textAlign","width","height","padding","margin","border","borderRadius","display","position","top","right","bottom","left","zIndex","flexDirection","justifyContent","alignItems","gap","opacity","visibility","overflow","boxShadow","transform"];function $u(s){if(typeof window>"u")return"";const r=(s.ownerDocument.defaultView??window).getComputedStyle(s),c=[];for(const u of sv){const f=u.replace(/([A-Z])/g,"-$1").toLowerCase(),d=r.getPropertyValue(f);d&&!Fp.has(d)&&c.push(`${f}: ${d}`)}return c.join("; ")}function av(s){if(!s)return;const r={},c=s.split(";").map(u=>u.trim()).filter(Boolean);for(const u of c){const f=u.indexOf(":");if(f>0){const d=u.slice(0,f).trim(),m=u.slice(f+1).trim();d&&m&&(r[d]=m)}}return Object.keys(r).length>0?r:void 0}function Uu(s){const r=[],c=s.getAttribute("role"),u=s.getAttribute("aria-label"),f=s.getAttribute("aria-describedby"),d=s.getAttribute("tabindex"),m=s.getAttribute("aria-hidden");return c&&r.push(`role="${c}"`),u&&r.push(`aria-label="${u}"`),f&&r.push(`aria-describedby="${f}"`),d&&r.push(`tabindex=${d}`),m==="true"&&r.push("aria-hidden"),s.matches("a, button, input, select, textarea, [tabindex]")&&r.push("focusable"),r.join(", ")}function cc(s){const r=[];let c=s;for(;c&&c.tagName.toLowerCase()!=="html";){const f=c.tagName.toLowerCase();let d=f;if(c.id)d=`${f}#${c.id}`;else if(c.className&&typeof c.className=="string"){const g=c.className.split(/\s+/).map(p=>p.replace(/[_][a-zA-Z0-9]{5,}.*$/,"")).find(p=>p.length>2);g&&(d=`${f}.${g}`)}const m=Ir(c);!c.parentElement&&m&&(d=`⟨shadow⟩ ${d}`),r.unshift(d),c=m}const u=Xl(s.ownerDocument);return(u?cc(u)+" > ⟨iframe⟩ ":"")+r.join(" > ")}var Wp="agentation-toolbar, [data-agentation-root], [data-feedback-toolbar], [data-annotation-popup], [data-annotation-marker]",rv=new Set(["DIV","SPAN","SECTION","ARTICLE","MAIN","ASIDE","HEADER","FOOTER","NAV"]);function ed(s,r){var f,d,m;let c=document.elementFromPoint(s,r);const u=new Set;for(;c&&!u.has(c);){u.add(c);const g=mc(c);let p;if(g){const j=Qr(c);s=(s-j.x)/j.sx,r=(r-j.y)/j.sy,p=(f=g.elementFromPoint)==null?void 0:f.call(g,s,r)}else p=(m=(d=c.shadowRoot)==null?void 0:d.elementFromPoint)==null?void 0:m.call(d,s,r);if(!p||p===c)break;c=p}return c}function iv(s){if(typeof s.checkVisibility=="function")return s.checkVisibility({checkOpacity:!0,checkVisibilityCSS:!0});const r=getComputedStyle(s);if(r.visibility==="hidden"||r.visibility==="collapse")return!1;let c=s;for(;c;){const u=getComputedStyle(c);if(u.opacity==="0"||u.display==="none"||u.contentVisibility==="hidden")return!1;const f=c.getRootNode();c=c.parentElement||(G_(f)?f.host:null)}return!0}function Kp(s,r){var d;const c=[],u=new Set,f=(m,g,p)=>{var j,k,b,v;for(const z of m){if(u.has(z))continue;if(u.add(z),z.shadowRoot){const $=z.shadowRoot,B=((j=$.elementsFromPoint)==null?void 0:j.call($,g,p))??[];f(B.length?B:[(k=$.elementFromPoint)==null?void 0:k.call($,g,p)].filter(Boolean),g,p)}const E=mc(z);if(E){const $=Qr(z),B=(g-$.x)/$.sx,F=(p-$.y)/$.sy;f(((b=E.elementsFromPoint)==null?void 0:b.call(E,B,F))??[(v=E.elementFromPoint)==null?void 0:v.call(E,B,F)].filter(Boolean),B,F)}z!==z.ownerDocument.body&&z!==z.ownerDocument.documentElement&&!ol(z,Wp)&&iv(z)&&c.push(z)}};return f(((d=document.elementsFromPoint)==null?void 0:d.call(document,s,r))??[document.elementFromPoint(s,r)].filter(Boolean),s,r),c}function Hu(s,r,c){if(c.width<=0||c.height<=0)return null;let u=null,f=1/0;for(const d of Kp(s,r)){const m=Un(d),g=m.width/c.width,p=m.height/c.height;if(g<.5||g>2||p<.5||p>2)continue;const j=Math.abs(Math.log(g))+Math.abs(Math.log(p));j<f&&(u=d,f=j)}return u}function B0(s,r){const c=ed(s,r);if(!c||ol(c,Wp))return null;const u=Kp(s,r);for(const m of u)if(!rv.has(m.tagName)&&!m.shadowRoot||Array.from(m.childNodes).some(g=>{var p;return g.nodeType===Node.TEXT_NODE&&((p=g.textContent)==null?void 0:p.trim())}))return m;let f=null,d=1/0;for(const m of u){const g=Un(m),p=g.width*g.height;p>0&&p<d&&(f=m,d=p)}return f}function cv(s){const r=h.useCallback(c=>s?(window.addEventListener("hashchange",c),window.addEventListener("popstate",c),()=>{window.removeEventListener("hashchange",c),window.removeEventListener("popstate",c)}):()=>{},[s]);return h.useSyncExternalStore(r,()=>window.location.pathname+(s?window.location.hash:""),()=>"/")}var Yu=new Map;function uv(s,r){const c=(Yu.get(s)??Promise.resolve()).then(r),u=c.then(()=>{},()=>{});return Yu.set(s,u),u.then(()=>{Yu.get(s)===u&&Yu.delete(s)}),c}function dv(s,r,c){try{const u=new URL(s,c);return u.origin===c&&u.pathname+u.hash===r}catch{return!1}}var $0=typeof window>"u"?h.useEffect:h.useLayoutEffect;function fv(s){const[r,c]=h.useState(null);return $0(()=>{const u=document.createElement("div");return u.setAttribute("data-agentation-portal",""),u.style.display="contents",c(u),()=>u.remove()},[]),$0(()=>{var m;if(!r)return;const u=s??document.body;if(u.ownerDocument!==document){console.warn("[Agentation] portalContainer belongs to another document; the toolbar will not render.");return}let f=document.activeElement;for(;(m=f==null?void 0:f.shadowRoot)!=null&&m.activeElement;)f=f.shadowRoot.activeElement;const d=f&&r.contains(document.activeElement)?f:null;typeof r.hidePopover=="function"&&r.matches(":popover-open")&&r.hidePopover(),u.appendChild(r),s&&typeof r.showPopover=="function"?(r.setAttribute("popover","manual"),r.style.cssText="position:fixed;inset:0 auto auto 0;margin:0;padding:0;border:0;background:transparent;width:0;height:0;overflow:visible;pointer-events:none",r.showPopover()):(r.removeAttribute("popover"),r.style.cssText="display:contents"),d==null||d.focus({preventScroll:!0})},[r,s]),r}var _v=({mode:s="open",delegatesFocus:r,slotAssignment:c,host:u="div",children:f,className:d,...m})=>{const g=h.useRef(null),[p,j]=h.useState(null);h.useLayoutEffect(()=>{const b=g.current;if(!b||b.shadowRoot)return;const v=b.attachShadow({mode:s,delegatesFocus:r,slotAssignment:c});j(v)},[]);const k=u;return o.jsx(k,{ref:g,...m,...u.includes("-")?{class:d}:{className:d},children:p&&L_.createPortal(f,p)})};function F_(s,r,c){const u=h.useRef(c);h.useLayoutEffect(()=>{u.current=c},[c]),h.useLayoutEffect(()=>{var g;const f=s.current;if(!r||!f)return;let d=!1;const m=((g=f.getAnimations)==null?void 0:g.call(f))??[];return Promise.allSettled(m.map(p=>p.finished)).then(()=>{d||u.current()}),()=>{d=!0}},[s,r])}var Zp=`@charset "UTF-8";
.styles-module__popup___IhzrD svg[fill=none] {
  fill: none !important;
}
.styles-module__popup___IhzrD svg[fill=none] :not([fill]) {
  fill: none !important;
}

@keyframes styles-module__popupEnter___AuQDN {
  from {
    opacity: 0;
    transform: translateX(-50%) scale(0.95) translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) scale(1) translateY(0);
  }
}
@keyframes styles-module__popupExit___JJKQX {
  from {
    opacity: 1;
    transform: translateX(-50%) scale(1) translateY(0);
  }
  to {
    opacity: 0;
    transform: translateX(-50%) scale(0.95) translateY(4px);
  }
}
@keyframes styles-module__shake___jdbWe {
  0%, 100% {
    transform: translateX(-50%) scale(1) translateY(0) translateX(0);
  }
  20% {
    transform: translateX(-50%) scale(1) translateY(0) translateX(-3px);
  }
  40% {
    transform: translateX(-50%) scale(1) translateY(0) translateX(3px);
  }
  60% {
    transform: translateX(-50%) scale(1) translateY(0) translateX(-2px);
  }
  80% {
    transform: translateX(-50%) scale(1) translateY(0) translateX(2px);
  }
}
.styles-module__popup___IhzrD {
  position: fixed;
  transform: translateX(-50%);
  width: 280px;
  padding: 0.75rem 1rem;
  background: #1a1a1a;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.08);
  z-index: 100001;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  will-change: transform, opacity;
  opacity: 0;
}
.styles-module__popup___IhzrD.styles-module__enter___L7U7N {
  animation: styles-module__popupEnter___AuQDN 0.2s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
}
.styles-module__popup___IhzrD.styles-module__entered___COX-w {
  opacity: 1;
  transform: translateX(-50%) scale(1) translateY(0);
}
.styles-module__popup___IhzrD.styles-module__exit___5eGjE {
  pointer-events: none;
  animation: styles-module__popupExit___JJKQX 0.15s ease-in forwards;
}
.styles-module__popup___IhzrD.styles-module__entered___COX-w.styles-module__shake___jdbWe {
  animation: styles-module__shake___jdbWe 0.25s ease-out;
}

.styles-module__header___wWsSi {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5625rem;
}

.styles-module__element___fTV2z {
  font-size: 0.75rem;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.5);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.styles-module__headerToggle___WpW0b {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  flex: 1;
  min-width: 0;
  text-align: left;
}
.styles-module__headerToggle___WpW0b .styles-module__element___fTV2z {
  flex: 1;
}

.styles-module__chevron___ZZJlR {
  color: rgba(255, 255, 255, 0.5);
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  flex-shrink: 0;
}
.styles-module__chevron___ZZJlR.styles-module__expanded___2Hxgv {
  transform: rotate(90deg);
}

.styles-module__stylesWrapper___pnHgy {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.styles-module__stylesWrapper___pnHgy.styles-module__expanded___2Hxgv {
  grid-template-rows: 1fr;
}

.styles-module__stylesInner___YYZe2 {
  overflow: hidden;
}

.styles-module__stylesBlock___VfQKn {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0.375rem;
  padding: 0.5rem 0.625rem;
  margin-bottom: 0.5rem;
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
  font-size: 0.6875rem;
  line-height: 1.5;
}

.styles-module__styleLine___1YQiD {
  color: rgba(255, 255, 255, 0.85);
  word-break: break-word;
}

.styles-module__styleProperty___84L1i {
  color: #c792ea;
}

.styles-module__styleValue___q51-h {
  color: rgba(255, 255, 255, 0.85);
}

.styles-module__timestamp___Dtpsv {
  font-size: 0.625rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.35);
  font-variant-numeric: tabular-nums;
  margin-left: 0.5rem;
  flex-shrink: 0;
}

.styles-module__quote___mcMmQ {
  font-size: 12px;
  font-style: italic;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 0.5rem;
  padding: 0.4rem 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0.25rem;
  line-height: 1.45;
}

.styles-module__textarea___jrSae {
  box-sizing: border-box;
  width: 100%;
  padding: 0.5rem 0.625rem;
  font-size: 0.8125rem;
  font-family: inherit;
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  resize: none;
  outline: none;
  transition: border-color 0.15s ease;
}
.styles-module__textarea___jrSae:focus {
  border-color: var(--agentation-color-blue);
}
.styles-module__textarea___jrSae.styles-module__green___99l3h:focus {
  border-color: var(--agentation-color-green);
}
.styles-module__textarea___jrSae::placeholder {
  color: rgba(255, 255, 255, 0.35);
}
.styles-module__textarea___jrSae::-webkit-scrollbar {
  width: 6px;
}
.styles-module__textarea___jrSae::-webkit-scrollbar-track {
  background: transparent;
}
.styles-module__textarea___jrSae::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.styles-module__actions___D6x3f {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.375rem;
  margin-top: 0.75rem;
}

.styles-module__sourceAction___EabJb {
  display: block;
  max-width: 100%;
  margin: -2px 0 8px;
  padding: 2px 0;
  background: transparent;
  color: #fff;
  font: inherit;
  font-size: 11px;
  border: 0;
  cursor: pointer;
  opacity: 0.6;
  transition: opacity 0.15s ease;
}
.styles-module__sourceAction___EabJb:hover, .styles-module__sourceAction___EabJb:focus-visible {
  opacity: 1;
}
.styles-module__sourceAction___EabJb:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 3px;
}

.styles-module__light___6AaSQ .styles-module__sourceAction___EabJb {
  color: #111;
}

.styles-module__cancel___hRjnL,
.styles-module__submit___K-mIR,
.styles-module__deleteButton___4VuAE {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.875rem;
  padding: 0.375rem 0.875rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 1rem;
  border: none;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease, opacity 0.15s ease;
}

.styles-module__cancel___hRjnL {
  background: transparent;
  color: rgba(255, 255, 255, 0.5);
}
.styles-module__cancel___hRjnL:hover {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}

.styles-module__submit___K-mIR {
  color: white;
}
.styles-module__submit___K-mIR:hover:not(:disabled) {
  filter: brightness(0.9);
}
.styles-module__submit___K-mIR:disabled {
  cursor: not-allowed;
}

.styles-module__deleteWrapper___oSjdo {
  display: flex;
  margin-right: auto;
}

.styles-module__deleteButton___4VuAE {
  background: transparent;
  color: rgba(255, 255, 255, 0.4);
  transition: background-color 0.15s ease, color 0.15s ease, transform 0.1s ease;
}
.styles-module__deleteButton___4VuAE:hover {
  background-color: color-mix(in srgb, var(--agentation-color-red) 25%, transparent);
  color: var(--agentation-color-red);
}
.styles-module__deleteButton___4VuAE:active {
  transform: scale(0.92);
}

.styles-module__light___6AaSQ.styles-module__popup___IhzrD {
  background: #fff;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(0, 0, 0, 0.06);
}
.styles-module__light___6AaSQ .styles-module__element___fTV2z {
  color: rgba(0, 0, 0, 0.6);
}
.styles-module__light___6AaSQ .styles-module__timestamp___Dtpsv {
  color: rgba(0, 0, 0, 0.4);
}
.styles-module__light___6AaSQ .styles-module__chevron___ZZJlR {
  color: rgba(0, 0, 0, 0.4);
}
.styles-module__light___6AaSQ .styles-module__stylesBlock___VfQKn {
  background: rgba(0, 0, 0, 0.03);
}
.styles-module__light___6AaSQ .styles-module__styleLine___1YQiD {
  color: rgba(0, 0, 0, 0.75);
}
.styles-module__light___6AaSQ .styles-module__styleProperty___84L1i {
  color: #7c3aed;
}
.styles-module__light___6AaSQ .styles-module__styleValue___q51-h {
  color: rgba(0, 0, 0, 0.75);
}
.styles-module__light___6AaSQ .styles-module__quote___mcMmQ {
  color: rgba(0, 0, 0, 0.55);
  background: rgba(0, 0, 0, 0.04);
}
.styles-module__light___6AaSQ .styles-module__textarea___jrSae {
  background: rgba(0, 0, 0, 0.03);
  color: #1a1a1a;
  border-color: rgba(0, 0, 0, 0.12);
}
.styles-module__light___6AaSQ .styles-module__textarea___jrSae::placeholder {
  color: rgba(0, 0, 0, 0.4);
}
.styles-module__light___6AaSQ .styles-module__textarea___jrSae::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
}
.styles-module__light___6AaSQ .styles-module__cancel___hRjnL {
  color: rgba(0, 0, 0, 0.5);
}
.styles-module__light___6AaSQ .styles-module__cancel___hRjnL:hover {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.75);
}
.styles-module__light___6AaSQ .styles-module__deleteButton___4VuAE {
  color: rgba(0, 0, 0, 0.4);
}
.styles-module__light___6AaSQ .styles-module__deleteButton___4VuAE:hover {
  background-color: color-mix(in srgb, var(--agentation-color-red) 25%, transparent);
  color: var(--agentation-color-red);
}

@media (prefers-reduced-motion: reduce) {
  .styles-module__popup___IhzrD.styles-module__enter___L7U7N, .styles-module__popup___IhzrD.styles-module__exit___5eGjE {
    animation-duration: 1ms;
    animation-delay: 0ms !important;
  }
}
.styles-module__sharedForm___8GvQl {
  --card-motion: 200ms cubic-bezier(0.2, 0.8, 0.2, 1);
  padding: 0.75rem 1rem;
  transition: padding var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__header___wWsSi {
  transition: margin-bottom var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__element___fTV2z {
  font-style: italic;
  color: rgba(255, 255, 255, 0.6);
}
.styles-module__sharedForm___8GvQl .styles-module__previewExcerpt___DCOIL {
  display: none;
}
.styles-module__sharedForm___8GvQl .styles-module__headerToggle___WpW0b .styles-module__chevron___ZZJlR {
  opacity: 1;
  margin-left: 0;
  transition: margin-left var(--card-motion), opacity 100ms ease-out, transform var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__sharedNote___OYVi5 {
  position: relative;
  height: var(--editor-field-height, 57px);
  border-radius: 8px;
  overflow: clip;
  transition: height var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__sharedNote___OYVi5::before {
  content: "";
  position: absolute;
  inset: 0;
  border: 1px solid var(--field-border, rgba(255, 255, 255, 0.15));
  border-radius: inherit;
  background: rgba(255, 255, 255, 0.05);
  pointer-events: none;
  transition: opacity var(--card-motion), border-color var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__sharedNoteContent___6q3KD {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: clip;
  transition: width var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__sharedNote___OYVi5[data-truncated]::after {
  content: "…";
  position: absolute;
  right: 0;
  top: 0;
  color: #fff;
  font-size: 13px;
  line-height: 1.4;
  opacity: 0;
  pointer-events: none;
  transition: opacity 80ms ease-out;
}
.styles-module__sharedForm___8GvQl .styles-module__textarea___jrSae {
  display: block;
  width: calc(280px - 2rem);
  max-width: calc(100vw - 24px - 2rem);
  margin: 0;
  background: transparent !important;
  border-color: transparent !important;
  transform: translate(0, 0);
  transition: transform var(--card-motion), color var(--card-motion);
}
.styles-module__sharedForm___8GvQl .styles-module__sharedExtra___RUBKC, .styles-module__sharedForm___8GvQl .styles-module__sharedActions___6Glpl {
  display: grid;
  grid-template-rows: 1fr;
  opacity: 1;
  transition: grid-template-rows var(--card-motion), opacity 80ms ease-out;
}
.styles-module__sharedForm___8GvQl .styles-module__sharedActions___6Glpl {
  transition-delay: 0ms, 120ms;
}
.styles-module__sharedForm___8GvQl .styles-module__sharedExtraInner___EuUh4 {
  min-height: 0;
  overflow: hidden;
}
.styles-module__sharedForm___8GvQl .styles-module__actions___D6x3f {
  min-height: 0;
  overflow: hidden;
  transition: margin-top var(--card-motion);
}
.styles-module__sharedForm___8GvQl[data-preview] {
  padding: 8px 12px;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__header___wWsSi {
  margin-bottom: 5px;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__previewExcerpt___DCOIL {
  display: inline;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__element___fTV2z {
  line-height: 1.4;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__chevron___ZZJlR {
  opacity: 0;
  margin-left: -18px;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedNote___OYVi5 {
  height: 20.2px;
  border-radius: 0;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedNote___OYVi5::before {
  opacity: 0;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__textarea___jrSae {
  transform: translate(-11px, calc(-9px + (1.4em - 1lh) / 2));
  color: #fff;
  overflow: hidden;
  cursor: default;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedExtra___RUBKC, .styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedActions___6Glpl {
  grid-template-rows: 0fr;
  opacity: 0;
  transition-delay: 0ms;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__actions___D6x3f {
  margin-top: 0;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__stylesWrapper___pnHgy {
  grid-template-rows: 0fr;
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedNote___OYVi5[data-truncated] .styles-module__sharedNoteContent___6q3KD {
  width: calc(100% - 12px);
}
.styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedNote___OYVi5[data-truncated]::after {
  opacity: 1;
}

.styles-module__light___6AaSQ .styles-module__sharedForm___8GvQl .styles-module__sharedNote___OYVi5::before {
  border-color: var(--field-border, rgba(0, 0, 0, 0.12));
  background: rgba(0, 0, 0, 0.03);
}

.styles-module__light___6AaSQ .styles-module__sharedForm___8GvQl .styles-module__element___fTV2z {
  color: rgba(0, 0, 0, 0.5);
}

.styles-module__light___6AaSQ .styles-module__sharedForm___8GvQl[data-preview] .styles-module__textarea___jrSae, .styles-module__light___6AaSQ .styles-module__sharedForm___8GvQl[data-preview] .styles-module__sharedNote___OYVi5::after {
  color: rgba(0, 0, 0, 0.85);
}

@media (prefers-reduced-motion: reduce) {
  .styles-module__sharedForm___8GvQl {
    --card-motion: 1ms linear;
  }
  .styles-module__sharedForm___8GvQl .styles-module__sharedActions___6Glpl, .styles-module__sharedForm___8GvQl .styles-module__headerToggle___WpW0b .styles-module__chevron___ZZJlR {
    transition: opacity 100ms ease-out;
    transition-delay: 0ms;
  }
}`,xt={popup:"styles-module__popup___IhzrD",enter:"styles-module__enter___L7U7N",entered:"styles-module__entered___COX-w",exit:"styles-module__exit___5eGjE",shake:"styles-module__shake___jdbWe",header:"styles-module__header___wWsSi",element:"styles-module__element___fTV2z",headerToggle:"styles-module__headerToggle___WpW0b",chevron:"styles-module__chevron___ZZJlR",expanded:"styles-module__expanded___2Hxgv",stylesWrapper:"styles-module__stylesWrapper___pnHgy",stylesInner:"styles-module__stylesInner___YYZe2",stylesBlock:"styles-module__stylesBlock___VfQKn",styleLine:"styles-module__styleLine___1YQiD",styleProperty:"styles-module__styleProperty___84L1i",styleValue:"styles-module__styleValue___q51-h",timestamp:"styles-module__timestamp___Dtpsv",quote:"styles-module__quote___mcMmQ",textarea:"styles-module__textarea___jrSae",actions:"styles-module__actions___D6x3f",sourceAction:"styles-module__sourceAction___EabJb",light:"styles-module__light___6AaSQ",cancel:"styles-module__cancel___hRjnL",submit:"styles-module__submit___K-mIR",deleteButton:"styles-module__deleteButton___4VuAE",deleteWrapper:"styles-module__deleteWrapper___oSjdo",sharedForm:"styles-module__sharedForm___8GvQl",previewExcerpt:"styles-module__previewExcerpt___DCOIL",sharedNote:"styles-module__sharedNote___OYVi5",sharedNoteContent:"styles-module__sharedNoteContent___6q3KD",sharedExtra:"styles-module__sharedExtra___RUBKC",sharedActions:"styles-module__sharedActions___6Glpl",sharedExtraInner:"styles-module__sharedExtraInner___EuUh4"},y_="data-agentation-styles";function Jp(s,r,c){if(!c||!s)return;const u=s.nodeType===9?s.head:s.nodeType===11?s:null;if(!u||typeof u.querySelector!="function"||u.querySelector(`style[${y_}~="toolbar"], style[${y_}~="${r}"]`))return;const d=(s.nodeType===9?s:s.ownerDocument).createElement("style");d.setAttribute(y_,r),d.textContent=c,u.appendChild(d)}function W_(s,r){return h.useCallback(c=>{c&&Jp(c.getRootNode(),s,r)},[s,r])}function U0(s){if(!s)return;const r=c=>c.stopImmediatePropagation();document.addEventListener("focusin",r,!0),document.addEventListener("focusout",r,!0);try{s.focus({preventScroll:!0})}finally{document.removeEventListener("focusin",r,!0),document.removeEventListener("focusout",r,!0)}}var Pp=h.forwardRef(function({element:r,timestamp:c,selectedText:u,placeholder:f="What should change?",initialValue:d="",submitLabel:m="Add",onSubmit:g,onCancel:p,onDelete:j,onOpenSource:k,allowEmpty:b=!1,accentColor:v="#3c82f7",computedStyles:z,disabled:E=!1,preview:$=!1,resetOnPreview:B=!0,variant:F="popup"},Me){const te=F==="card",[C,ie]=h.useState(d),[de,W]=h.useState(!1),[Ee,ne]=h.useState(!1),Ce=h.useRef(null),pe=h.useRef(null),V=u?` "${u.slice(0,30)}${u.length>30?"...":""}"`:"";h.useLayoutEffect(()=>{var Oe;const ue=pe.current,G=Ce.current;if(!te||!ue||!G)return;const we=()=>{ue.style.setProperty("--editor-field-height",`${G.offsetHeight}px`)};if(we(),"CanvasRenderingContext2D"in window){const Se=document.createElement("canvas").getContext("2d");if(Se){const Qe=getComputedStyle(G).fontFamily;Se.font=`13px ${Qe}`;const Ut=Se.measureText(d.replace(/\s+/g," ")).width;Se.font=`italic 12px ${Qe}`;const Ye=Se.measureText(r+V).width,Ne=Math.min(200,Math.max(120,Math.ceil(Math.max(Ut,Ye))+24));(Oe=ue.closest("[data-annotation-card]"))==null||Oe.style.setProperty("--preview-width",`${Ne}px`);const Xt=ue.querySelector("[data-shared-note]");Xt&&Xt.toggleAttribute("data-truncated",Ut>Ne-24)}}const fe=typeof ResizeObserver<"u"?new ResizeObserver(we):null;return fe==null||fe.observe(G),()=>fe==null?void 0:fe.disconnect()},[te,r,d,V]),h.useLayoutEffect(()=>{$&&B&&(ie(d),ne(!1),Ce.current&&(Ce.current.scrollTop=0,Ce.current.scrollLeft=0))},[$,B,d]),h.useImperativeHandle(Me,()=>({focus(){const ue=Ce.current;U0(ue),ue&&(ue.selectionStart=ue.selectionEnd=ue.value.length,ue.scrollTop=te?0:ue.scrollHeight)}}),[te]);const Ie=h.useCallback(()=>{E||!C.trim()&&!b||g(C.trim())},[E,C,b,g]),he=ue=>{ue.stopPropagation(),!ue.nativeEvent.isComposing&&(ue.key==="Enter"&&!ue.shiftKey&&(ue.preventDefault(),Ie()),ue.key==="Escape"&&p())};return o.jsxs("div",{ref:pe,className:te?xt.sharedForm:void 0,style:te?void 0:{display:"contents"},"data-annotation-editor":!0,"data-preview":$||void 0,children:[o.jsxs("div",{className:xt.header,"data-editor-heading":!0,children:[z&&Object.keys(z).length>0?o.jsxs("button",{className:xt.headerToggle,onClick:()=>{const ue=Ee;ne(!Ee),ue&&Ht(()=>U0(Ce.current),0)},type:"button",children:[o.jsx("svg",{className:`${xt.chevron} ${Ee?xt.expanded:""}`,width:"14",height:"14",viewBox:"0 0 14 14",fill:"none",xmlns:"http://www.w3.org/2000/svg",children:o.jsx("path",{d:"M5.5 10.25L9 7.25L5.75 4",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})}),o.jsxs("span",{className:xt.element,children:[r,te&&V&&o.jsx("span",{className:xt.previewExcerpt,children:V})]})]}):o.jsxs("span",{className:xt.element,children:[r,te&&V&&o.jsx("span",{className:xt.previewExcerpt,children:V})]}),c&&o.jsx("span",{className:xt.timestamp,children:c})]}),k&&o.jsx("div",{className:te?xt.sharedExtra:void 0,style:te?void 0:{display:"contents"},children:o.jsx("div",{className:te?xt.sharedExtraInner:void 0,style:te?void 0:{display:"contents"},children:o.jsx("button",{type:"button",className:xt.sourceAction,onClick:k,children:"Open in editor"})})}),z&&Object.keys(z).length>0&&o.jsx("div",{className:`${xt.stylesWrapper} ${Ee?xt.expanded:""}`,children:o.jsx("div",{className:xt.stylesInner,children:o.jsx("div",{className:xt.stylesBlock,children:Object.entries(z).map(([ue,G])=>o.jsxs("div",{className:xt.styleLine,children:[o.jsx("span",{className:xt.styleProperty,children:ue.replace(/([A-Z])/g,"-$1").toLowerCase()}),": ",o.jsx("span",{className:xt.styleValue,children:G}),";"]},ue))})})}),u&&o.jsx("div",{className:te?xt.sharedExtra:void 0,style:te?void 0:{display:"contents"},children:o.jsx("div",{className:te?xt.sharedExtraInner:void 0,style:te?void 0:{display:"contents"},children:o.jsxs("div",{className:xt.quote,children:["“",u.slice(0,80),u.length>80?"...":"","”"]})})}),o.jsx("div",{"data-shared-note":!0,className:te?xt.sharedNote:void 0,style:te?{"--field-border":de?v:void 0}:{display:"contents"},children:o.jsx("div",{className:te?xt.sharedNoteContent:void 0,style:te?void 0:{display:"contents"},children:o.jsx("textarea",{ref:Ce,className:xt.textarea,readOnly:$,"aria-hidden":$,style:te?void 0:{borderColor:de?v:void 0},placeholder:f,value:$?C.replace(/\s+/g," "):C,onChange:ue=>ie(ue.target.value),onFocus:()=>W(!0),onBlur:()=>W(!1),rows:2,onKeyDown:he})})}),o.jsx("div",{"data-editor-actions":!0,className:te?xt.sharedActions:void 0,style:te?void 0:{display:"contents"},children:o.jsxs("div",{className:xt.actions,children:[j&&o.jsx("div",{className:xt.deleteWrapper,children:o.jsx("button",{className:xt.deleteButton,onClick:j,type:"button","aria-label":"Delete annotation",children:"Delete"})}),o.jsx("button",{className:xt.cancel,onClick:p,children:"Cancel"}),o.jsx("button",{className:xt.submit,style:{backgroundColor:v,opacity:C.trim()||b?1:.4},onClick:Ie,disabled:E||!C.trim()&&!b,children:m})]})})]})}),K_=h.forwardRef(function({element:r,timestamp:c,selectedText:u,placeholder:f="What should change?",initialValue:d="",submitLabel:m="Add",onSubmit:g,onCancel:p,onDelete:j,onOpenSource:k,allowEmpty:b=!1,style:v,accentColor:z="#3c82f7",isExiting:E=!1,onExitComplete:$,lightMode:B=!1,computedStyles:F},Me){const[te,C]=h.useState(!1),[ie,de]=h.useState("initial"),W=h.useRef(null),Ee=h.useRef(null);h.useEffect(()=>{var he;Jp((he=Ee.current)==null?void 0:he.getRootNode(),"annotation-popup",Zp)},[]);const ne=h.useRef(null);h.useEffect(()=>{const he=Ht(()=>{de(ue=>ue==="initial"?"enter":ue)},0);return()=>{clearTimeout(he),ne.current&&clearTimeout(ne.current)}},[]),h.useEffect(()=>{if(E)return;const he=Ht(()=>{var ue;return(ue=W.current)==null?void 0:ue.focus()},50);return()=>clearTimeout(he)},[E]);const Ce=h.useCallback(()=>{ne.current&&clearTimeout(ne.current),C(!0),ne.current=Ht(()=>{var he;C(!1),(he=W.current)==null||he.focus()},250)},[]);h.useImperativeHandle(Me,()=>({shake:Ce}),[Ce]);const pe=h.useCallback(()=>{if($){p();return}de("exit")},[p,$]),V=E?"exit":ie;F_(Ee,V==="exit",()=>{E?$==null||$():p()});const Ie=[xt.popup,B?xt.light:"",V==="enter"?xt.enter:"",V==="entered"?xt.entered:"",V==="exit"?xt.exit:"",te&&V!=="exit"?xt.shake:""].filter(Boolean).join(" ");return o.jsx("div",{ref:Ee,className:Ie,"data-annotation-popup":!0,style:v,onAnimationEnd:he=>{he.target===he.currentTarget&&he.animationName.includes("popupEnter")&&!E&&de("entered")},onKeyDownCapture:he=>{he.key!=="Escape"||he.nativeEvent.isComposing||(he.preventDefault(),he.stopPropagation(),pe())},onClick:he=>he.stopPropagation(),children:o.jsx(Pp,{ref:W,element:r,timestamp:c,selectedText:u,placeholder:f,initialValue:d,submitLabel:m,onSubmit:g,onCancel:pe,onDelete:j,onOpenSource:k,allowEmpty:b,accentColor:z,computedStyles:F,disabled:V==="exit"})})}),sd=`.icon-transitions-module__iconState___uqK9J {
  transition: opacity 0.2s ease, transform 0.2s ease;
  transform-origin: center;
}

.icon-transitions-module__iconStateFast___HxlMm {
  transition: opacity 0.15s ease, transform 0.15s ease;
  transform-origin: center;
}

.icon-transitions-module__iconFade___nPwXg {
  transition: opacity 0.2s ease;
}

.icon-transitions-module__iconFadeFast___Ofb2t {
  transition: opacity 0.15s ease;
}

.icon-transitions-module__visible___PlHsU {
  opacity: 1 !important;
}

.icon-transitions-module__visibleScaled___8Qog- {
  opacity: 1 !important;
  transform: scale(1);
}

.icon-transitions-module__hidden___ETykt {
  opacity: 0 !important;
}

.icon-transitions-module__hiddenScaled___JXn-m {
  opacity: 0 !important;
  transform: scale(0.8);
}

.icon-transitions-module__sending___uaLN- {
  opacity: 0.5 !important;
  transform: scale(0.8);
}`,nn={iconState:"icon-transitions-module__iconState___uqK9J",iconStateFast:"icon-transitions-module__iconStateFast___HxlMm",iconFade:"icon-transitions-module__iconFade___nPwXg",iconFadeFast:"icon-transitions-module__iconFadeFast___Ofb2t",visible:"icon-transitions-module__visible___PlHsU",visibleScaled:"icon-transitions-module__visibleScaled___8Qog-",hidden:"icon-transitions-module__hidden___ETykt",hiddenScaled:"icon-transitions-module__hiddenScaled___JXn-m",sending:"icon-transitions-module__sending___uaLN-"},mv=({size:s=16})=>o.jsx("svg",{width:s,height:s,viewBox:"0 0 16 16",fill:"none",children:o.jsx("path",{d:"M8 3v10M3 8h10",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round"})}),hv=({size:s=20,...r})=>o.jsxs("svg",{width:s,height:s,viewBox:"0 0 20 20",fill:"none",xmlns:"http://www.w3.org/2000/svg",...r,children:[o.jsx("circle",{cx:"10",cy:"10",r:"5.375",stroke:"currentColor",strokeWidth:"1.25"}),o.jsx("path",{d:"M8.5 8.5C8.73 7.85 9.31 7.49 10 7.5C10.86 7.51 11.5 8.13 11.5 9C11.5 10.08 10 10.5 10 10.5V10.75",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("circle",{cx:"10",cy:"12.625",r:"0.625",fill:"currentColor"})]}),gv=({size:s=24,copied:r=!1,tint:c})=>o.jsxs("svg",{ref:W_("icon-transitions",sd),width:s,height:s,viewBox:"0 0 24 24",fill:"none",style:c?{color:c,transition:"color 0.3s ease"}:void 0,children:[o.jsxs("g",{className:`${nn.iconState} ${r?nn.hiddenScaled:nn.visibleScaled}`,children:[o.jsx("path",{d:"M4.75 11.25C4.75 10.4216 5.42157 9.75 6.25 9.75H12.75C13.5784 9.75 14.25 10.4216 14.25 11.25V17.75C14.25 18.5784 13.5784 19.25 12.75 19.25H6.25C5.42157 19.25 4.75 18.5784 4.75 17.75V11.25Z",stroke:"currentColor",strokeWidth:"1.5"}),o.jsx("path",{d:"M17.25 14.25H17.75C18.5784 14.25 19.25 13.5784 19.25 12.75V6.25C19.25 5.42157 18.5784 4.75 17.75 4.75H11.25C10.4216 4.75 9.75 5.42157 9.75 6.25V6.75",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round"})]}),o.jsxs("g",{className:`${nn.iconState} ${r?nn.visibleScaled:nn.hiddenScaled}`,children:[o.jsx("path",{d:"M12 20C7.58172 20 4 16.4182 4 12C4 7.58172 7.58172 4 12 4C16.4182 4 20 7.58172 20 12C20 16.4182 16.4182 20 12 20Z",stroke:"var(--agentation-color-green)",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M15 10L11 14.25L9.25 12.25",stroke:"var(--agentation-color-green)",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})]})]}),pv=({size:s=24,state:r="idle"})=>{const c=r==="idle",u=r==="sent",f=r==="failed",d=r==="sending";return o.jsxs("svg",{width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:[o.jsx("g",{className:`${nn.iconStateFast} ${c?nn.visibleScaled:d?nn.sending:nn.hiddenScaled}`,children:o.jsx("path",{d:"M9.875 14.125L12.3506 19.6951C12.7184 20.5227 13.9091 20.4741 14.2083 19.6193L18.8139 6.46032C19.0907 5.6695 18.3305 4.90933 17.5397 5.18611L4.38072 9.79174C3.52589 10.0909 3.47731 11.2816 4.30494 11.6494L9.875 14.125ZM9.875 14.125L13.375 10.625",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})}),o.jsxs("g",{className:`${nn.iconStateFast} ${u?nn.visibleScaled:nn.hiddenScaled}`,children:[o.jsx("path",{d:"M12 20C7.58172 20 4 16.4182 4 12C4 7.58172 7.58172 4 12 4C16.4182 4 20 7.58172 20 12C20 16.4182 16.4182 20 12 20Z",stroke:"var(--agentation-color-green)",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M15 10L11 14.25L9.25 12.25",stroke:"var(--agentation-color-green)",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})]}),o.jsxs("g",{className:`${nn.iconStateFast} ${f?nn.visibleScaled:nn.hiddenScaled}`,children:[o.jsx("path",{d:"M12 20C7.58172 20 4 16.4182 4 12C4 7.58172 7.58172 4 12 4C16.4182 4 20 7.58172 20 12C20 16.4182 16.4182 20 12 20Z",stroke:"var(--agentation-color-red)",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M12 8V12",stroke:"var(--agentation-color-red)",strokeWidth:"1.5",strokeLinecap:"round"}),o.jsx("circle",{cx:"12",cy:"15",r:"0.5",fill:"var(--agentation-color-red)",stroke:"var(--agentation-color-red)",strokeWidth:"1"})]})]})},yv=({size:s=24,isOpen:r=!0})=>o.jsxs("svg",{ref:W_("icon-transitions",sd),width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:[o.jsxs("g",{className:`${nn.iconFade} ${r?nn.visible:nn.hidden}`,children:[o.jsx("path",{d:"M3.91752 12.7539C3.65127 12.2996 3.65037 11.7515 3.9149 11.2962C4.9042 9.59346 7.72688 5.49994 12 5.49994C16.2731 5.49994 19.0958 9.59346 20.0851 11.2962C20.3496 11.7515 20.3487 12.2996 20.0825 12.7539C19.0908 14.4459 16.2694 18.4999 12 18.4999C7.73064 18.4999 4.90918 14.4459 3.91752 12.7539Z",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M12 14.8261C13.5608 14.8261 14.8261 13.5608 14.8261 12C14.8261 10.4392 13.5608 9.17392 12 9.17392C10.4392 9.17392 9.17391 10.4392 9.17391 12C9.17391 13.5608 10.4392 14.8261 12 14.8261Z",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})]}),o.jsxs("g",{className:`${nn.iconFade} ${r?nn.hidden:nn.visible}`,children:[o.jsx("path",{d:"M18.6025 9.28503C18.9174 8.9701 19.4364 8.99481 19.7015 9.35271C20.1484 9.95606 20.4943 10.507 20.7342 10.9199C21.134 11.6086 21.1329 12.4454 20.7303 13.1328C20.2144 14.013 19.2151 15.5225 17.7723 16.8193C16.3293 18.1162 14.3852 19.2497 12.0008 19.25C11.4192 19.25 10.8638 19.1823 10.3355 19.0613C9.77966 18.934 9.63498 18.2525 10.0382 17.8493C10.2412 17.6463 10.5374 17.573 10.8188 17.6302C11.1993 17.7076 11.5935 17.75 12.0008 17.75C13.8848 17.7497 15.4867 16.8568 16.7693 15.7041C18.0522 14.5511 18.9606 13.1867 19.4363 12.375C19.5656 12.1543 19.5659 11.8943 19.4373 11.6729C19.2235 11.3049 18.921 10.8242 18.5364 10.3003C18.3085 9.98991 18.3302 9.5573 18.6025 9.28503ZM12.0008 4.75C12.5814 4.75006 13.1358 4.81803 13.6632 4.93953C14.2182 5.06741 14.362 5.74812 13.9593 6.15091C13.7558 6.35435 13.4589 6.42748 13.1771 6.36984C12.7983 6.29239 12.4061 6.25006 12.0008 6.25C10.1167 6.25 8.51415 7.15145 7.23028 8.31543C5.94678 9.47919 5.03918 10.8555 4.56426 11.6729C4.43551 11.8945 4.43582 12.1542 4.56524 12.375C4.77587 12.7343 5.07189 13.2012 5.44718 13.7105C5.67623 14.0213 5.65493 14.4552 5.38193 14.7282C5.0671 15.0431 4.54833 15.0189 4.28292 14.6614C3.84652 14.0736 3.50813 13.5369 3.27129 13.1328C2.86831 12.4451 2.86717 11.6088 3.26739 10.9199C3.78185 10.0345 4.77959 8.51239 6.22247 7.2041C7.66547 5.89584 9.61202 4.75 12.0008 4.75Z",fill:"currentColor"}),o.jsx("path",{d:"M5 19L19 5",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round"})]})]}),xv=({size:s=24,isPaused:r=!1})=>o.jsxs("svg",{ref:W_("icon-transitions",sd),width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:[o.jsxs("g",{className:`${nn.iconFadeFast} ${r?nn.hidden:nn.visible}`,children:[o.jsx("path",{d:"M8 6L8 18",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round"}),o.jsx("path",{d:"M16 18L16 6",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round"})]}),o.jsx("path",{className:`${nn.iconFadeFast} ${r?nn.visible:nn.hidden}`,d:"M17.75 10.701C18.75 11.2783 18.75 12.7217 17.75 13.299L8.75 18.4952C7.75 19.0725 6.5 18.3509 6.5 17.1962L6.5 6.80384C6.5 5.64914 7.75 4.92746 8.75 5.50481L17.75 10.701Z",stroke:"currentColor",strokeWidth:"1.5"})]}),bv=({size:s=16})=>o.jsxs("svg",{width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:[o.jsx("path",{d:"M10.6504 5.81117C10.9939 4.39628 13.0061 4.39628 13.3496 5.81117C13.5715 6.72517 14.6187 7.15891 15.4219 6.66952C16.6652 5.91193 18.0881 7.33479 17.3305 8.57815C16.8411 9.38134 17.2748 10.4285 18.1888 10.6504C19.6037 10.9939 19.6037 13.0061 18.1888 13.3496C17.2748 13.5715 16.8411 14.6187 17.3305 15.4219C18.0881 16.6652 16.6652 18.0881 15.4219 17.3305C14.6187 16.8411 13.5715 17.2748 13.3496 18.1888C13.0061 19.6037 10.9939 19.6037 10.6504 18.1888C10.4285 17.2748 9.38135 16.8411 8.57815 17.3305C7.33479 18.0881 5.91193 16.6652 6.66952 15.4219C7.15891 14.6187 6.72517 13.5715 5.81117 13.3496C4.39628 13.0061 4.39628 10.9939 5.81117 10.6504C6.72517 10.4285 7.15891 9.38134 6.66952 8.57815C5.91193 7.33479 7.33479 5.91192 8.57815 6.66952C9.38135 7.15891 10.4285 6.72517 10.6504 5.81117Z",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("circle",{cx:"12",cy:"12",r:"2.5",stroke:"currentColor",strokeWidth:"1.5"})]}),vv=({size:s=16})=>o.jsx("svg",{width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:o.jsx("path",{d:"M13.5 4C14.7426 4 15.75 5.00736 15.75 6.25V7H18.5C18.9142 7 19.25 7.33579 19.25 7.75C19.25 8.16421 18.9142 8.5 18.5 8.5H17.9678L17.6328 16.2217C17.61 16.7475 17.5912 17.1861 17.5469 17.543C17.5015 17.9087 17.4225 18.2506 17.2461 18.5723C16.9747 19.0671 16.5579 19.4671 16.0518 19.7168C15.7227 19.8791 15.3772 19.9422 15.0098 19.9717C14.6514 20.0004 14.2126 20 13.6865 20H10.3135C9.78735 20 9.34856 20.0004 8.99023 19.9717C8.62278 19.9422 8.27729 19.8791 7.94824 19.7168C7.44205 19.4671 7.02532 19.0671 6.75391 18.5723C6.57751 18.2506 6.49853 17.9087 6.45312 17.543C6.40883 17.1861 6.39005 16.7475 6.36719 16.2217L6.03223 8.5H5.5C5.08579 8.5 4.75 8.16421 4.75 7.75C4.75 7.33579 5.08579 7 5.5 7H8.25V6.25C8.25 5.00736 9.25736 4 10.5 4H13.5ZM7.86621 16.1562C7.89013 16.7063 7.90624 17.0751 7.94141 17.3584C7.97545 17.6326 8.02151 17.7644 8.06934 17.8516C8.19271 18.0763 8.38239 18.2577 8.6123 18.3711C8.70153 18.4151 8.83504 18.4545 9.11035 18.4766C9.39482 18.4994 9.76335 18.5 10.3135 18.5H13.6865C14.2367 18.5 14.6052 18.4994 14.8896 18.4766C15.165 18.4545 15.2985 18.4151 15.3877 18.3711C15.6176 18.2577 15.8073 18.0763 15.9307 17.8516C15.9785 17.7644 16.0245 17.6326 16.0586 17.3584C16.0938 17.0751 16.1099 16.7063 16.1338 16.1562L16.4668 8.5H7.5332L7.86621 16.1562ZM9.97656 10.75C10.3906 10.7371 10.7371 11.0626 10.75 11.4766L10.875 15.4766C10.8879 15.8906 10.5624 16.2371 10.1484 16.25C9.73443 16.2629 9.38794 15.9374 9.375 15.5234L9.25 11.5234C9.23706 11.1094 9.56255 10.7629 9.97656 10.75ZM14.0244 10.75C14.4384 10.7635 14.7635 11.1105 14.75 11.5244L14.6201 15.5244C14.6066 15.9384 14.2596 16.2634 13.8457 16.25C13.4317 16.2365 13.1067 15.8896 13.1201 15.4756L13.251 11.4756C13.2645 11.0617 13.6105 10.7366 14.0244 10.75ZM10.5 5.5C10.0858 5.5 9.75 5.83579 9.75 6.25V7H14.25V6.25C14.25 5.83579 13.9142 5.5 13.5 5.5H10.5Z",fill:"currentColor"})}),wv=({size:s=16})=>o.jsxs("svg",{width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:[o.jsxs("g",{clipPath:"url(#clip0_2_53)",children:[o.jsx("path",{d:"M16.25 16.25L7.75 7.75",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M7.75 16.25L16.25 7.75",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})]}),o.jsx("defs",{children:o.jsx("clipPath",{id:"clip0_2_53",children:o.jsx("rect",{width:"24",height:"24",fill:"white"})})})]}),Sv=({size:s=16})=>o.jsxs("svg",{width:s,height:s,viewBox:"0 0 20 20",fill:"none",children:[o.jsx("path",{d:"M9.99999 12.7082C11.4958 12.7082 12.7083 11.4956 12.7083 9.99984C12.7083 8.50407 11.4958 7.2915 9.99999 7.2915C8.50422 7.2915 7.29166 8.50407 7.29166 9.99984C7.29166 11.4956 8.50422 12.7082 9.99999 12.7082Z",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M10 3.9585V5.05698",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M10 14.9429V16.0414",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M5.7269 5.72656L6.50682 6.50649",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M13.4932 13.4932L14.2731 14.2731",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M3.95834 10H5.05683",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M14.9432 10H16.0417",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M5.7269 14.2731L6.50682 13.4932",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"}),o.jsx("path",{d:"M13.4932 6.50649L14.2731 5.72656",stroke:"currentColor",strokeWidth:"1.25",strokeLinecap:"round",strokeLinejoin:"round"})]}),jv=({size:s=16})=>o.jsx("svg",{width:s,height:s,viewBox:"0 0 20 20",fill:"none",children:o.jsx("path",{d:"M15.5 10.4955C15.4037 11.5379 15.0124 12.5314 14.3721 13.3596C13.7317 14.1878 12.8688 14.8165 11.8841 15.1722C10.8995 15.5278 9.83397 15.5957 8.81217 15.3679C7.79038 15.1401 6.8546 14.6259 6.11434 13.8857C5.37408 13.1454 4.85995 12.2096 4.63211 11.1878C4.40427 10.166 4.47215 9.10048 4.82781 8.11585C5.18346 7.13123 5.81218 6.26825 6.64039 5.62791C7.4686 4.98756 8.46206 4.59634 9.5045 4.5C8.89418 5.32569 8.60049 6.34302 8.67685 7.36695C8.75321 8.39087 9.19454 9.35339 9.92058 10.0794C10.6466 10.8055 11.6091 11.2468 12.6331 11.3231C13.657 11.3995 14.6743 11.1058 15.5 10.4955Z",stroke:"currentColor",strokeWidth:"1.13793",strokeLinecap:"round",strokeLinejoin:"round"})}),kv=({size:s=16})=>o.jsx("svg",{width:s,height:s,viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg",children:o.jsx("path",{d:"M11.3799 6.9572L9.05645 4.63375M11.3799 6.9572L6.74949 11.5699C6.61925 11.6996 6.45577 11.791 6.277 11.8339L4.29549 12.3092C3.93194 12.3964 3.60478 12.0683 3.69297 11.705L4.16585 9.75693C4.20893 9.57947 4.29978 9.4172 4.42854 9.28771L9.05645 4.63375M11.3799 6.9572L12.3455 5.98759C12.9839 5.34655 12.9839 4.31002 12.3455 3.66897C11.7033 3.02415 10.6594 3.02415 10.0172 3.66897L9.06126 4.62892L9.05645 4.63375",stroke:"currentColor",strokeWidth:"0.9",strokeLinecap:"round",strokeLinejoin:"round"})}),Cv=({size:s=16})=>o.jsx("svg",{width:s,height:s,viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg",children:o.jsx("path",{d:"M8.5 3.5L4 8L8.5 12.5",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})}),Nv=({size:s=24})=>o.jsxs("svg",{width:s,height:s,viewBox:"0 0 24 24",fill:"none",children:[o.jsx("rect",{x:"3",y:"3",width:"18",height:"18",rx:"2",stroke:"currentColor",strokeWidth:"1.5"}),o.jsx("line",{x1:"3",y1:"9",x2:"21",y2:"9",stroke:"currentColor",strokeWidth:"1.5"}),o.jsx("line",{x1:"9",y1:"9",x2:"9",y2:"21",stroke:"currentColor",strokeWidth:"1.5"})]}),Mv=({content:s,children:r,...c})=>{const[u,f]=h.useState(!1),[d,m]=h.useState(!1),[g,p]=h.useState({top:0,right:0}),j=h.useRef(null),k=h.useRef(null),b=h.useRef(null),v=()=>{if(j.current){const $=j.current.getBoundingClientRect();p({top:$.top+$.height/2,right:window.innerWidth-$.left+8})}},z=()=>{m(!0),b.current&&(clearTimeout(b.current),b.current=null),v(),k.current=Ht(()=>{f(!0)},500)},E=()=>{k.current&&(clearTimeout(k.current),k.current=null),f(!1),b.current=Ht(()=>{m(!1)},150)};return h.useEffect(()=>()=>{k.current&&clearTimeout(k.current),b.current&&clearTimeout(b.current)},[]),o.jsxs(o.Fragment,{children:[o.jsx("span",{ref:j,onMouseEnter:z,onMouseLeave:E,...c,children:r}),d&&L_.createPortal(o.jsx("div",{"data-feedback-toolbar":!0,style:{position:"fixed",top:g.top,right:g.right,transform:"translateY(-50%)",padding:"6px 10px",background:"#383838",color:"rgba(255, 255, 255, 0.7)",fontSize:"11px",fontWeight:400,lineHeight:"14px",borderRadius:"10px",width:"180px",textAlign:"left",zIndex:100020,pointerEvents:"none",boxShadow:"0px 1px 8px rgba(0, 0, 0, 0.28)",opacity:u?1:0,transition:"opacity 0.15s ease"},children:s}),document.body)]})},Ev=`.styles-module__tooltip___mcXL2 {
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: help;
}

.styles-module__tooltipIcon___Nq2nD {
  transform: translateY(0.5px);
  color: #fff;
  opacity: 0.2;
  transition: opacity 0.15s ease;
  will-change: transform;
}
.styles-module__tooltip___mcXL2:hover .styles-module__tooltipIcon___Nq2nD {
  opacity: 0.5;
}
[data-agentation-theme=light] .styles-module__tooltipIcon___Nq2nD {
  color: #000;
}`,H0={tooltip:"styles-module__tooltip___mcXL2",tooltipIcon:"styles-module__tooltipIcon___Nq2nD"},La=({content:s})=>o.jsx(Mv,{className:H0.tooltip,content:s,children:o.jsx(hv,{className:H0.tooltipIcon})}),Tv=`.styles-module__toolbar___wNsdK svg[fill=none],
.styles-module__markersLayer___-25j1 svg[fill=none],
.styles-module__fixedMarkersLayer___ffyX6 svg[fill=none] {
  fill: none !important;
}
.styles-module__toolbar___wNsdK svg[fill=none] :not([fill]),
.styles-module__markersLayer___-25j1 svg[fill=none] :not([fill]),
.styles-module__fixedMarkersLayer___ffyX6 svg[fill=none] :not([fill]) {
  fill: none !important;
}

.styles-module__controlsContent___9GJWU :where(button, input, select, textarea, label) {
  background: unset;
  border: unset;
  border-radius: unset;
  padding: unset;
  margin: unset;
  color: unset;
  font-family: unset;
  font-weight: unset;
  font-style: unset;
  line-height: unset;
  letter-spacing: unset;
  text-transform: unset;
  text-decoration: unset;
  box-shadow: unset;
  outline: unset;
}

@keyframes styles-module__toolbarEnter___u8RRu {
  from {
    opacity: 0;
    transform: scale(0.5) rotate(90deg);
  }
  to {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}
@keyframes styles-module__toolbarHide___y8kaT {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.8);
  }
}
@keyframes styles-module__badgeEnter___mVQLj {
  from {
    opacity: 0;
    transform: scale(0);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
@keyframes styles-module__scaleIn___c-r1K {
  from {
    opacity: 0;
    transform: scale(0.85);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
@keyframes styles-module__scaleOut___Wctwz {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.85);
  }
}
@keyframes styles-module__slideUp___kgD36 {
  from {
    opacity: 0;
    transform: scale(0.85) translateY(8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
@keyframes styles-module__slideDown___zcdje {
  from {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
  to {
    opacity: 0;
    transform: scale(0.85) translateY(8px);
  }
}
@keyframes styles-module__fadeIn___b9qmf {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
@keyframes styles-module__fadeOut___6Ut6- {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}
@keyframes styles-module__hoverHighlightIn___6WYHY {
  from {
    opacity: 0;
    transform: scale(0.98);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
@keyframes styles-module__hoverTooltipIn___FYGQx {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(4px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
.styles-module__disableTransitions___EopxO :is(*, *::before, *::after) {
  transition: none !important;
}

:host {
  /* Set here rather than inline so a consumer className rule can still hide the toolbar. */
  display: contents;
  position: fixed;
  top: auto;
  left: auto;
  bottom: 1.25rem;
  right: 1.25rem;
  z-index: 100000;
}

.styles-module__positionContext___AZFHE,
.styles-module__toolbar___wNsdK {
  position: inherit;
  top: inherit;
  left: inherit;
  bottom: inherit;
  right: inherit;
  z-index: inherit;
}

.styles-module__toolbar___wNsdK {
  width: 337px;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  pointer-events: none;
  transition: left 0.36s cubic-bezier(0.19, 1, 0.22, 1), top 0s, right 0s, bottom 0s;
}
.styles-module__toolbar___wNsdK[data-dragging=true] {
  transition: none;
}

.styles-module__toolbarContainer___dIhma {
  position: relative;
  -webkit-user-select: none;
  user-select: none;
  margin-left: auto;
  align-self: flex-end;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1a1a1a;
  color: #fff;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2), 0 4px 16px rgba(0, 0, 0, 0.1);
  pointer-events: auto;
  transition: width 0.36s cubic-bezier(0.19, 1, 0.22, 1), transform 0.36s cubic-bezier(0.19, 1, 0.22, 1);
}
.styles-module__toolbarContainer___dIhma.styles-module__entrance___sgHd8 {
  animation: styles-module__toolbarEnter___u8RRu 0.5s cubic-bezier(0.34, 1.2, 0.64, 1) forwards;
}
.styles-module__toolbarContainer___dIhma.styles-module__hiding___1td44 {
  animation: styles-module__toolbarHide___y8kaT 0.4s cubic-bezier(0.4, 0, 1, 1) forwards;
  pointer-events: none;
}
.styles-module__toolbarContainer___dIhma.styles-module__collapsed___Rydsn {
  width: 44px;
  height: 44px;
  border-radius: 22px;
  padding: 0;
  cursor: pointer;
}
.styles-module__toolbarContainer___dIhma.styles-module__collapsed___Rydsn:hover {
  background: #2a2a2a;
}
.styles-module__toolbarContainer___dIhma.styles-module__collapsed___Rydsn:active {
  transform: scale(0.95);
}
.styles-module__toolbarContainer___dIhma.styles-module__expanded___ofKPx {
  height: 44px;
  border-radius: 22px;
  padding: 5px;
  width: 297px;
}
.styles-module__toolbarContainer___dIhma.styles-module__expanded___ofKPx.styles-module__serverConnected___Gfbou {
  width: 337px;
}

@media (prefers-reduced-motion: reduce) {
  .styles-module__toolbar___wNsdK,
  .styles-module__toolbarContainer___dIhma {
    transition: none;
  }
}
.styles-module__buttonWrapper___rBcdv.styles-module__toggleWrapper___7N0-q {
  position: absolute;
  top: 0;
  right: 0;
  width: 44px;
  height: 44px;
}

.styles-module__togglePlaceholder___wnqrL {
  width: 34px;
  flex: 0 0 34px;
  height: 34px;
}

.styles-module__toggleContent___0yfyP {
  position: absolute;
  top: 0;
  right: 0;
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  padding: 0;
  margin: 0;
  background: transparent;
  color: #fff;
  cursor: pointer;
  transition: color 0.15s ease;
}
.styles-module__toggleContent___0yfyP::before {
  content: "";
  position: absolute;
  inset: 5px;
  border-radius: 50%;
  pointer-events: none;
  background: transparent;
  transition: background-color 0.15s ease, transform 0.1s ease;
}
.styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN {
  color: rgba(255, 255, 255, 0.85);
}
.styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN:hover {
  color: #fff;
}
.styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN:hover::before {
  background: rgba(255, 255, 255, 0.12);
}
.styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN:active::before, .styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN:active .styles-module__toggleGlyph___R7Oom {
  transform: scale(0.92);
}

.styles-module__toggleIcon___Jbtus {
  transform: translateY(-0.5px);
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1);
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.styles-module__expandedToggle___F7SRN .styles-module__toggleIcon___Jbtus {
  transform: none;
}

.styles-module__toggleGlyph___R7Oom {
  overflow: visible;
  transition: transform 0.1s ease;
}
.styles-module__toggleGlyph___R7Oom path {
  transform-box: fill-box;
  transform-origin: center;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.16s ease;
}

.styles-module__toggleTopLine___hQaCm,
.styles-module__toggleMiddleLine___sFFVe {
  vector-effect: non-scaling-stroke;
}

.styles-module__toggleBottomLine___V-jX3 {
  transform-origin: left center;
}

.styles-module__toggleGlyph___R7Oom[data-active=true] .styles-module__toggleTopLine___hQaCm {
  transform: translateY(5.25px) rotate(45deg) scaleX(1.1422494);
}
.styles-module__toggleGlyph___R7Oom[data-active=true] .styles-module__toggleMiddleLine___sFFVe {
  transform: translateX(3.5px) rotate(-45deg) scaleX(2.4748737);
}
.styles-module__toggleGlyph___R7Oom[data-active=true] .styles-module__toggleBottomLine___V-jX3 {
  transform: scaleX(0);
  opacity: 0;
}
.styles-module__toggleGlyph___R7Oom[data-active=true] .styles-module__toggleSparkle___eeF99 {
  transform: scale(0);
  opacity: 0;
}

.styles-module__toggleContent___0yfyP:focus-visible,
.styles-module__controlButton___8Q0jc:focus-visible {
  outline: 2px solid var(--agentation-color-accent);
  outline-offset: 3px;
}

.styles-module__controlsContent___9GJWU {
  display: flex;
  align-items: center;
  gap: 6px;
  transition: filter 0.14s ease-out, opacity 0.14s ease-out, transform 0.36s cubic-bezier(0.19, 1, 0.22, 1);
}
.styles-module__controlsContent___9GJWU.styles-module__visible___KHwEW {
  transition: filter 0.3s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.24s ease-out, transform 0.42s cubic-bezier(0.19, 1, 0.22, 1);
  opacity: 1;
  filter: blur(0px);
  transform: scale(1);
  visibility: visible;
  pointer-events: auto;
}
.styles-module__controlsContent___9GJWU.styles-module__hidden___Ae8H4 {
  pointer-events: none;
  opacity: 0;
  filter: blur(6px);
  transform: scale(0.4);
}

@media (prefers-reduced-motion: reduce) {
  .styles-module__controlsContent___9GJWU,
  .styles-module__controlsContent___9GJWU.styles-module__visible___KHwEW,
  .styles-module__toggleContent___0yfyP,
  .styles-module__toggleContent___0yfyP::before,
  .styles-module__toggleGlyph___R7Oom,
  .styles-module__toggleIcon___Jbtus,
  .styles-module__toggleGlyph___R7Oom path {
    transition: none;
  }
  .styles-module__controlsContent___9GJWU.styles-module__hidden___Ae8H4 {
    filter: none;
    transform: none;
  }
}
.styles-module__badge___2XsgF {
  position: absolute;
  top: -13px;
  right: -13px;
  -webkit-user-select: none;
  user-select: none;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background-color: var(--agentation-color-accent);
  color: white;
  font-size: 0.625rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15), inset 0 0 0 1px rgba(255, 255, 255, 0.04);
  opacity: 1;
  transition: transform 0.3s ease, opacity 0.2s ease;
  transform: scale(1);
}
.styles-module__badge___2XsgF.styles-module__fadeOut___6Ut6- {
  opacity: 0;
  transform: scale(0);
  pointer-events: none;
}
.styles-module__badge___2XsgF.styles-module__entrance___sgHd8 {
  animation: styles-module__badgeEnter___mVQLj 0.3s cubic-bezier(0.34, 1.2, 0.64, 1) 0.4s both;
}

.styles-module__controlButton___8Q0jc {
  position: relative;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.85);
  transition: background-color 0.15s ease, color 0.15s ease, transform 0.1s ease, opacity 0.2s ease;
}
.styles-module__controlButton___8Q0jc:hover:not(:disabled):not([data-active=true]):not([data-failed=true]):not([data-auto-sync=true]):not([data-error=true]):not([data-no-hover=true]) {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}
.styles-module__controlButton___8Q0jc:active:not(:disabled) {
  transform: scale(0.92);
}
.styles-module__controlButton___8Q0jc:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.styles-module__controlButton___8Q0jc[data-active=true] {
  color: var(--agentation-color-blue);
  background-color: color-mix(in srgb, var(--agentation-color-blue) 25%, transparent);
}
.styles-module__controlButton___8Q0jc[data-error=true] {
  color: var(--agentation-color-red);
  background-color: color-mix(in srgb, var(--agentation-color-red) 25%, transparent);
}
.styles-module__controlButton___8Q0jc[data-danger]:hover:not(:disabled):not([data-active=true]):not([data-failed=true]) {
  background-color: color-mix(in srgb, var(--agentation-color-red) 25%, transparent);
  color: var(--agentation-color-red);
}
.styles-module__controlButton___8Q0jc[data-no-hover=true], .styles-module__controlButton___8Q0jc.styles-module__statusShowing___te6iu {
  cursor: default;
  pointer-events: none;
  background: transparent !important;
}
.styles-module__controlButton___8Q0jc[data-auto-sync=true] {
  color: var(--agentation-color-green);
  background: transparent;
  cursor: default;
}
.styles-module__controlButton___8Q0jc[data-failed=true] {
  color: var(--agentation-color-red);
  background-color: color-mix(in srgb, var(--agentation-color-red) 25%, transparent);
}

.styles-module__buttonBadge___NeFWb {
  position: absolute;
  top: 0px;
  right: 0px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background-color: var(--agentation-color-accent);
  color: white;
  font-size: 0.625rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 0 2px #1a1a1a, 0 1px 3px rgba(0, 0, 0, 0.2);
  pointer-events: none;
}
[data-agentation-theme=light] .styles-module__buttonBadge___NeFWb {
  box-shadow: 0 0 0 2px #fff, 0 1px 3px rgba(0, 0, 0, 0.2);
}

@keyframes styles-module__mcpIndicatorPulseConnected___EDodZ {
  0%, 100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-green) 50%, transparent);
  }
  50% {
    box-shadow: 0 0 0 5px color-mix(in srgb, var(--agentation-color-green) 0%, transparent);
  }
}
@keyframes styles-module__mcpIndicatorPulseConnecting___cCYte {
  0%, 100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-yellow) 50%, transparent);
  }
  50% {
    box-shadow: 0 0 0 5px color-mix(in srgb, var(--agentation-color-yellow) 0%, transparent);
  }
}
.styles-module__mcpIndicator___zGJeL {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  pointer-events: none;
  transition: background-color 0.3s ease, opacity 0.15s ease, transform 0.15s ease;
  opacity: 1;
  transform: scale(1);
}
.styles-module__mcpIndicator___zGJeL.styles-module__connected___7c28g {
  background-color: var(--agentation-color-green);
  animation: styles-module__mcpIndicatorPulseConnected___EDodZ 2.5s ease-in-out infinite;
}
.styles-module__mcpIndicator___zGJeL.styles-module__connecting___uo-CW {
  background-color: var(--agentation-color-yellow);
  animation: styles-module__mcpIndicatorPulseConnecting___cCYte 1.5s ease-in-out infinite;
}
.styles-module__mcpIndicator___zGJeL.styles-module__hidden___Ae8H4 {
  opacity: 0;
  transform: scale(0);
  animation: none;
}

@keyframes styles-module__connectionPulse___-Zycw {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(0.9);
  }
}
.styles-module__connectionIndicatorWrapper___L-e-3 {
  width: 8px;
  height: 34px;
  margin-left: 6px;
  margin-right: 6px;
}

.styles-module__connectionIndicator___afk9p {
  position: relative;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  opacity: 0;
  transition: opacity 0.3s ease, background-color 0.3s ease;
  cursor: default;
}

.styles-module__connectionIndicatorVisible___C-i5B {
  opacity: 1;
}

.styles-module__connectionIndicatorConnected___IY8pR {
  background-color: var(--agentation-color-green);
  animation: styles-module__connectionPulse___-Zycw 2.5s ease-in-out infinite;
}

.styles-module__connectionIndicatorDisconnected___kmpaZ {
  background-color: var(--agentation-color-red);
  animation: none;
}

.styles-module__connectionIndicatorConnecting___QmSLH {
  background-color: var(--agentation-color-yellow);
  animation: styles-module__connectionPulse___-Zycw 1s ease-in-out infinite;
}

.styles-module__buttonWrapper___rBcdv {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.styles-module__buttonWrapper___rBcdv:hover .styles-module__buttonTooltip___Burd9 {
  opacity: 1;
  visibility: visible;
  transform: translateX(-50%) scale(1);
  transition-delay: 0.85s;
}
.styles-module__buttonWrapper___rBcdv:has(.styles-module__controlButton___8Q0jc:disabled):hover .styles-module__buttonTooltip___Burd9 {
  opacity: 0;
  visibility: hidden;
}

.styles-module__tooltipsInSession___-0lHH .styles-module__buttonWrapper___rBcdv:hover .styles-module__buttonTooltip___Burd9 {
  transition-delay: 0s;
}

.styles-module__sendButtonWrapper___UUxG6 {
  width: 0;
  opacity: 0;
  overflow: hidden;
  pointer-events: none;
  margin-left: -6px;
  transition: width 0.36s cubic-bezier(0.19, 1, 0.22, 1), opacity 0.2s cubic-bezier(0.19, 1, 0.22, 1), margin 0.36s cubic-bezier(0.19, 1, 0.22, 1);
}
.styles-module__sendButtonWrapper___UUxG6 .styles-module__controlButton___8Q0jc {
  transform: scale(0.8);
  transition: transform 0.36s cubic-bezier(0.19, 1, 0.22, 1);
}
.styles-module__sendButtonWrapper___UUxG6.styles-module__sendButtonVisible___WPSQU {
  width: 34px;
  opacity: 1;
  overflow: visible;
  pointer-events: auto;
  margin-left: 0;
}
.styles-module__sendButtonWrapper___UUxG6.styles-module__sendButtonVisible___WPSQU .styles-module__controlButton___8Q0jc {
  transform: scale(1);
}

.styles-module__buttonTooltip___Burd9 {
  position: absolute;
  bottom: calc(100% + 14px);
  left: 50%;
  transform: translateX(-50%) scale(0.95);
  padding: 6px 10px;
  background: #1a1a1a;
  color: rgba(255, 255, 255, 0.9);
  font-size: 12px;
  font-weight: 500;
  border-radius: 8px;
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  z-index: 100001;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  transition: opacity 0.135s ease, transform 0.135s ease, visibility 0.135s ease;
}
.styles-module__buttonTooltip___Burd9::after {
  content: "";
  position: absolute;
  top: calc(100% - 4px);
  left: 50%;
  transform: translateX(-50%) rotate(45deg);
  width: 8px;
  height: 8px;
  background: #1a1a1a;
  border-radius: 0 0 2px 0;
}

.styles-module__shortcut___lEAQk {
  margin-left: 4px;
  opacity: 0.5;
}

.styles-module__tooltipBelow___m6ats .styles-module__buttonTooltip___Burd9 {
  bottom: auto;
  top: calc(100% + 14px);
  transform: translateX(-50%) scale(0.95);
}
.styles-module__tooltipBelow___m6ats .styles-module__buttonTooltip___Burd9::after {
  top: -4px;
  bottom: auto;
  border-radius: 2px 0 0 0;
}

.styles-module__tooltipBelow___m6ats .styles-module__buttonWrapper___rBcdv:hover .styles-module__buttonTooltip___Burd9 {
  transform: translateX(-50%) scale(1);
}

.styles-module__tooltipsHidden___VtLJG .styles-module__buttonTooltip___Burd9 {
  opacity: 0 !important;
  visibility: hidden !important;
  transition: none !important;
}

.styles-module__tooltipVisible___0jcCv,
.styles-module__tooltipsHidden___VtLJG .styles-module__tooltipVisible___0jcCv {
  opacity: 1 !important;
  visibility: visible !important;
  transform: translateX(-50%) scale(1) !important;
  transition-delay: 0s !important;
}

.styles-module__buttonWrapperAlignLeft___myzIp .styles-module__buttonTooltip___Burd9 {
  left: 50%;
  transform: translateX(-12px) scale(0.95);
}
.styles-module__buttonWrapperAlignLeft___myzIp .styles-module__buttonTooltip___Burd9::after {
  left: 16px;
}
.styles-module__buttonWrapperAlignLeft___myzIp:hover .styles-module__buttonTooltip___Burd9 {
  transform: translateX(-12px) scale(1);
}

.styles-module__tooltipBelow___m6ats .styles-module__buttonWrapperAlignLeft___myzIp .styles-module__buttonTooltip___Burd9 {
  transform: translateX(-12px) scale(0.95);
}
.styles-module__tooltipBelow___m6ats .styles-module__buttonWrapperAlignLeft___myzIp:hover .styles-module__buttonTooltip___Burd9 {
  transform: translateX(-12px) scale(1);
}

.styles-module__buttonWrapperAlignRight___HCQFR .styles-module__buttonTooltip___Burd9 {
  left: 50%;
  transform: translateX(calc(-100% + 12px)) scale(0.95);
}
.styles-module__buttonWrapperAlignRight___HCQFR .styles-module__buttonTooltip___Burd9::after {
  left: auto;
  right: 8px;
}
.styles-module__buttonWrapperAlignRight___HCQFR:hover .styles-module__buttonTooltip___Burd9 {
  transform: translateX(calc(-100% + 12px)) scale(1);
}

.styles-module__tooltipBelow___m6ats .styles-module__buttonWrapperAlignRight___HCQFR .styles-module__buttonTooltip___Burd9 {
  transform: translateX(calc(-100% + 12px)) scale(0.95);
}
.styles-module__tooltipBelow___m6ats .styles-module__buttonWrapperAlignRight___HCQFR:hover .styles-module__buttonTooltip___Burd9 {
  transform: translateX(calc(-100% + 12px)) scale(1);
}

.styles-module__divider___c--s1 {
  width: 1px;
  height: 12px;
  background: rgba(255, 255, 255, 0.15);
  margin: 0 3px;
}

.styles-module__overlay___Q1O9y {
  position: fixed;
  inset: 0;
  z-index: 99997;
  pointer-events: none;
}
.styles-module__overlay___Q1O9y > * {
  pointer-events: auto;
}

.styles-module__hoverHighlight___ogakW {
  position: fixed;
  border: 2px solid color-mix(in srgb, var(--agentation-color-accent) 50%, transparent);
  border-radius: 4px;
  background-color: color-mix(in srgb, var(--agentation-color-accent) 4%, transparent);
  pointer-events: none !important;
  box-sizing: border-box;
  will-change: opacity;
  contain: layout style;
}
.styles-module__hoverHighlight___ogakW.styles-module__enter___WFIki {
  animation: styles-module__hoverHighlightIn___6WYHY 0.12s ease-out forwards;
}

.styles-module__multiSelectOutline___cSJ-m {
  position: fixed;
  border: 2px dashed color-mix(in srgb, var(--agentation-color-green) 60%, transparent);
  border-radius: 4px;
  pointer-events: none !important;
  background-color: color-mix(in srgb, var(--agentation-color-green) 5%, transparent);
  box-sizing: border-box;
  will-change: opacity;
}
.styles-module__multiSelectOutline___cSJ-m.styles-module__enter___WFIki {
  animation: styles-module__fadeIn___b9qmf 0.15s ease-out forwards;
}
.styles-module__multiSelectOutline___cSJ-m.styles-module__exit___fyOJ0 {
  animation: styles-module__fadeOut___6Ut6- 0.15s ease-out forwards;
}

.styles-module__singleSelectOutline___QhX-O {
  position: fixed;
  border: 2px solid color-mix(in srgb, var(--agentation-color-blue) 60%, transparent);
  border-radius: 4px;
  pointer-events: none !important;
  background-color: color-mix(in srgb, var(--agentation-color-blue) 5%, transparent);
  box-sizing: border-box;
  will-change: opacity;
}
.styles-module__singleSelectOutline___QhX-O.styles-module__enter___WFIki {
  animation: styles-module__fadeIn___b9qmf 0.15s ease-out forwards;
}
.styles-module__singleSelectOutline___QhX-O.styles-module__exit___fyOJ0 {
  animation: styles-module__fadeOut___6Ut6- 0.15s ease-out forwards;
}

.styles-module__hoverTooltip___bvLk7 {
  position: fixed;
  z-index: 99999;
  font-size: 0.6875rem;
  font-weight: 500;
  color: #fff;
  background: rgba(0, 0, 0, 0.85);
  padding: 0.35rem 0.6rem;
  border-radius: 0.375rem;
  pointer-events: none !important;
  white-space: nowrap;
  max-width: min(280px, 100vw - 16px - 1.2rem);
  overflow: hidden;
  text-overflow: ellipsis;
}
.styles-module__hoverTooltip___bvLk7.styles-module__enter___WFIki {
  animation: styles-module__hoverTooltipIn___FYGQx 0.1s ease-out forwards;
}

.styles-module__hoverReactPath___gx1IJ {
  font-size: 0.625rem;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 0.15rem;
  overflow: hidden;
  text-overflow: ellipsis;
}

.styles-module__hoverElementName___QMLMl {
  overflow: hidden;
  text-overflow: ellipsis;
}

.styles-module__markersLayer___-25j1 {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 0;
  z-index: 99998;
  pointer-events: none;
}
.styles-module__markersLayer___-25j1 > * {
  pointer-events: auto;
}

.styles-module__fixedMarkersLayer___ffyX6 {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99998;
  pointer-events: none;
}
.styles-module__fixedMarkersLayer___ffyX6 > * {
  pointer-events: auto;
}

.styles-module__marker___6sQrs {
  position: absolute;
  width: 22px;
  height: 22px;
  background: var(--agentation-color-blue);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6875rem;
  font-weight: 600;
  transform: translate(-50%, -50%) scale(1);
  opacity: 1;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2), inset 0 0 0 1px rgba(0, 0, 0, 0.04);
  -webkit-user-select: none;
  user-select: none;
  will-change: transform, opacity;
  contain: layout style;
  z-index: 1;
}
.styles-module__marker___6sQrs:hover {
  z-index: 2;
}
.styles-module__marker___6sQrs:not(.styles-module__enter___WFIki):not(.styles-module__exit___fyOJ0):not(.styles-module__clearing___FQ--7) {
  transition: background-color 0.15s ease, transform 0.1s ease;
}
.styles-module__marker___6sQrs.styles-module__enter___WFIki {
  animation: styles-module__markerIn___5FaAP 0.25s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.styles-module__marker___6sQrs.styles-module__exit___fyOJ0 {
  animation: styles-module__markerOut___GU5jX 0.2s ease-out both;
  pointer-events: none;
}
.styles-module__marker___6sQrs.styles-module__clearing___FQ--7 {
  animation: styles-module__markerOut___GU5jX 0.15s ease-out both;
  pointer-events: none;
}
.styles-module__marker___6sQrs:not(.styles-module__enter___WFIki):not(.styles-module__exit___fyOJ0):not(.styles-module__clearing___FQ--7):hover {
  transform: translate(-50%, -50%) scale(1.1);
}
.styles-module__marker___6sQrs.styles-module__pending___2IHLC {
  position: fixed;
  background-color: var(--agentation-color-blue);
  cursor: default;
}
.styles-module__marker___6sQrs.styles-module__fixed___dBMHC {
  position: fixed;
}
.styles-module__marker___6sQrs.styles-module__multiSelect___YWiuz {
  background-color: var(--agentation-color-green);
  width: 26px;
  height: 26px;
  border-radius: 6px;
  font-size: 0.75rem;
}
.styles-module__marker___6sQrs.styles-module__multiSelect___YWiuz.styles-module__pending___2IHLC {
  background-color: var(--agentation-color-green);
}
.styles-module__marker___6sQrs.styles-module__hovered___ZgXIy {
  background-color: var(--agentation-color-red);
}

.styles-module__renumber___nCTxD {
  display: block;
  animation: styles-module__renumberRoll___Wgbq3 0.2s ease-out;
}

@keyframes styles-module__renumberRoll___Wgbq3 {
  0% {
    transform: translateX(-40%);
    opacity: 0;
  }
  100% {
    transform: translateX(0);
    opacity: 1;
  }
}
.styles-module__markerTooltip___aLJID {
  position: absolute;
  top: calc(100% + 10px);
  left: 50%;
  transform: translateX(-50%) scale(0.909);
  z-index: 100002;
  background: #1a1a1a;
  padding: 8px 0.75rem;
  border-radius: 0.75rem;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-weight: 400;
  color: #fff;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.08);
  min-width: 120px;
  max-width: 200px;
  pointer-events: none;
  cursor: default;
}
.styles-module__markerTooltip___aLJID.styles-module__enter___WFIki {
  animation: styles-module__tooltipIn___0N31w 0.1s ease-out forwards;
}

.styles-module__markerQuote___FHmrz {
  display: block;
  font-size: 12px;
  font-style: italic;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 0.3125rem;
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.styles-module__markerNote___QkrrS {
  display: block;
  font-size: 13px;
  font-weight: 400;
  line-height: 1.4;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-bottom: 2px;
}

.styles-module__markerHint___2iF-6 {
  display: block;
  font-size: 0.625rem;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 0.375rem;
  white-space: nowrap;
}

.styles-module__settingsPanel___OxX3Y {
  position: absolute;
  right: 5px;
  bottom: calc(100% + 0.5rem);
  z-index: 1;
  overflow: hidden;
  background: #1c1c1c;
  border-radius: 1rem;
  padding: 13px 0 16px;
  min-width: 205px;
  cursor: default;
  opacity: 1;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(0, 0, 0, 0.04);
  transition: background-color 0.25s ease, box-shadow 0.25s ease;
}
.styles-module__settingsPanel___OxX3Y::before, .styles-module__settingsPanel___OxX3Y::after {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  width: 16px;
  z-index: 2;
  pointer-events: none;
}
.styles-module__settingsPanel___OxX3Y::before {
  left: 0;
  background: linear-gradient(to right, #1c1c1c 0%, transparent 100%);
}
.styles-module__settingsPanel___OxX3Y::after {
  right: 0;
  background: linear-gradient(to left, #1c1c1c 0%, transparent 100%);
}
.styles-module__settingsPanel___OxX3Y .styles-module__settingsHeader___pwDY9,
.styles-module__settingsPanel___OxX3Y .styles-module__settingsBrand___0gJeM,
.styles-module__settingsPanel___OxX3Y .styles-module__settingsBrandSlash___uTG18,
.styles-module__settingsPanel___OxX3Y .styles-module__settingsVersion___TUcFq,
.styles-module__settingsPanel___OxX3Y .styles-module__settingsSection___m-YM2,
.styles-module__settingsPanel___OxX3Y .styles-module__settingsLabel___8UjfX,
.styles-module__settingsPanel___OxX3Y .styles-module__cycleButton___FMKfw,
.styles-module__settingsPanel___OxX3Y .styles-module__cycleDot___nPgLY,
.styles-module__settingsPanel___OxX3Y .styles-module__dropdownButton___16NPz,
.styles-module__settingsPanel___OxX3Y .styles-module__toggleLabel___Xm8Aa,
.styles-module__settingsPanel___OxX3Y .styles-module__customCheckbox___U39ax,
.styles-module__settingsPanel___OxX3Y .styles-module__sliderLabel___U8sPr,
.styles-module__settingsPanel___OxX3Y .styles-module__slider___GLdxp,
.styles-module__settingsPanel___OxX3Y .styles-module__themeToggle___2rUjA {
  transition: background-color 0.25s ease, color 0.25s ease, border-color 0.25s ease;
}
.styles-module__settingsPanel___OxX3Y.styles-module__enter___WFIki {
  opacity: 1;
  transform: translateY(0) scale(1);
  filter: blur(0px);
  transition: opacity 0.2s ease, transform 0.2s ease, filter 0.2s ease;
}
.styles-module__settingsPanel___OxX3Y.styles-module__exit___fyOJ0 {
  opacity: 0;
  transform: translateY(8px) scale(0.95);
  filter: blur(5px);
  pointer-events: none;
  transition: opacity 0.1s ease, transform 0.1s ease, filter 0.1s ease;
}
[data-agentation-theme=dark] .styles-module__settingsPanel___OxX3Y {
  background: #1a1a1a;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.08);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___OxX3Y .styles-module__settingsLabel___8UjfX {
  color: rgba(255, 255, 255, 0.6);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___OxX3Y .styles-module__settingsOption___UNa12 {
  color: rgba(255, 255, 255, 0.85);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___OxX3Y .styles-module__settingsOption___UNa12:hover {
  background: rgba(255, 255, 255, 0.1);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___OxX3Y .styles-module__settingsOption___UNa12.styles-module__selected___OwRqP {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}
[data-agentation-theme=dark] .styles-module__settingsPanel___OxX3Y .styles-module__toggleLabel___Xm8Aa {
  color: rgba(255, 255, 255, 0.85);
}

.styles-module__settingsPanelContainer___Xksv8 {
  overflow: visible;
  position: relative;
  display: flex;
  padding: 0 1rem;
}

.styles-module__settingsPage___6YfHH {
  min-width: 100%;
  flex-shrink: 0;
  transition: transform 0.2s ease, opacity 0.2s ease;
  transition-delay: 0s;
  opacity: 1;
}

.styles-module__settingsPage___6YfHH.styles-module__slideLeft___Ps01J {
  transform: translateX(-24px);
  opacity: 0;
  pointer-events: none;
}

.styles-module__automationsPage___uvCq6 {
  position: absolute;
  top: 0;
  left: 24px;
  width: 100%;
  height: 100%;
  padding: 3px 1rem 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  transition: transform 0.2s ease, opacity 0.2s ease;
  opacity: 0;
  pointer-events: none;
}

.styles-module__automationsPage___uvCq6.styles-module__slideIn___4-qXe {
  transform: translateX(-24px);
  opacity: 1;
  pointer-events: auto;
}

.styles-module__settingsNavLink___wCzJt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0;
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 0.8125rem;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: color 0.15s ease;
}
.styles-module__settingsNavLink___wCzJt:hover {
  color: rgba(255, 255, 255, 0.9);
}
[data-agentation-theme=light] .styles-module__settingsNavLink___wCzJt {
  color: rgba(0, 0, 0, 0.5);
}
[data-agentation-theme=light] .styles-module__settingsNavLink___wCzJt:hover {
  color: rgba(0, 0, 0, 0.8);
}
.styles-module__settingsNavLink___wCzJt svg {
  color: rgba(255, 255, 255, 0.4);
  transition: color 0.15s ease;
}
.styles-module__settingsNavLink___wCzJt:hover svg {
  color: #fff;
}
[data-agentation-theme=light] .styles-module__settingsNavLink___wCzJt svg {
  color: rgba(0, 0, 0, 0.25);
}
[data-agentation-theme=light] .styles-module__settingsNavLink___wCzJt:hover svg {
  color: rgba(0, 0, 0, 0.8);
}

.styles-module__settingsNavLinkRight___ZWwhj {
  display: flex;
  align-items: center;
  gap: 6px;
}

.styles-module__mcpNavIndicator___cl9pO {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.styles-module__mcpNavIndicator___cl9pO.styles-module__connected___7c28g {
  background-color: var(--agentation-color-green);
  animation: styles-module__mcpPulse___uNggr 2.5s ease-in-out infinite;
}
.styles-module__mcpNavIndicator___cl9pO.styles-module__connecting___uo-CW {
  background-color: var(--agentation-color-yellow);
  animation: styles-module__mcpPulse___uNggr 1.5s ease-in-out infinite;
}

.styles-module__settingsBackButton___bIe2j {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 0 12px 0;
  margin: -6px 0 0.5rem 0;
  border: none;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 0;
  background: transparent;
  font-family: inherit;
  font-size: 0.8125rem;
  font-weight: 500;
  letter-spacing: -0.15px;
  color: #fff;
  cursor: pointer;
  transition: transform 0.12s cubic-bezier(0.32, 0.72, 0, 1);
}
.styles-module__settingsBackButton___bIe2j svg {
  opacity: 0.4;
  flex-shrink: 0;
  transition: opacity 0.15s ease, transform 0.18s cubic-bezier(0.32, 0.72, 0, 1);
}
.styles-module__settingsBackButton___bIe2j:hover {
  border-bottom-color: rgba(255, 255, 255, 0.07);
}
.styles-module__settingsBackButton___bIe2j:hover svg {
  opacity: 1;
}
[data-agentation-theme=light] .styles-module__settingsBackButton___bIe2j {
  color: rgba(0, 0, 0, 0.85);
  border-bottom-color: rgba(0, 0, 0, 0.08);
}
[data-agentation-theme=light] .styles-module__settingsBackButton___bIe2j:hover {
  border-bottom-color: rgba(0, 0, 0, 0.08);
}

.styles-module__automationHeader___InP0r {
  display: flex;
  align-items: center;
  gap: 0.125rem;
  font-size: 0.8125rem;
  font-weight: 400;
  color: #fff;
}
[data-agentation-theme=light] .styles-module__automationHeader___InP0r {
  color: rgba(0, 0, 0, 0.85);
}

.styles-module__automationDescription___NKlmo {
  font-size: 0.6875rem;
  font-weight: 300;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 2px;
  line-height: 14px;
}
[data-agentation-theme=light] .styles-module__automationDescription___NKlmo {
  color: rgba(0, 0, 0, 0.5);
}

.styles-module__learnMoreLink___8xv-x {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: underline dotted;
  text-decoration-color: rgba(255, 255, 255, 0.2);
  text-underline-offset: 2px;
  transition: color 0.15s ease;
}
.styles-module__learnMoreLink___8xv-x:hover {
  color: #fff;
}
[data-agentation-theme=light] .styles-module__learnMoreLink___8xv-x {
  color: rgba(0, 0, 0, 0.6);
  text-decoration-color: rgba(0, 0, 0, 0.2);
}
[data-agentation-theme=light] .styles-module__learnMoreLink___8xv-x:hover {
  color: rgba(0, 0, 0, 0.85);
}

.styles-module__autoSendRow___UblX5 {
  display: flex;
  align-items: center;
  gap: 8px;
}

.styles-module__autoSendLabel___icDc2 {
  font-size: 0.6875rem;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.4);
  transition: color 0.15s ease;
}
.styles-module__autoSendLabel___icDc2.styles-module__active___-zoN6 {
  color: #66b8ff;
  color: color(display-p3 0.4 0.72 1);
}
[data-agentation-theme=light] .styles-module__autoSendLabel___icDc2 {
  color: rgba(0, 0, 0, 0.4);
}
[data-agentation-theme=light] .styles-module__autoSendLabel___icDc2.styles-module__active___-zoN6 {
  color: var(--agentation-color-blue);
}

.styles-module__webhookUrlInput___2375C {
  display: block;
  width: 100%;
  flex: 1;
  min-height: 60px;
  box-sizing: border-box;
  margin-top: 11px;
  padding: 8px 10px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.03);
  font-family: inherit;
  font-size: 0.75rem;
  font-weight: 400;
  color: #fff;
  outline: none;
  resize: none;
  user-select: text;
  transition: border-color 0.15s ease, background-color 0.15s ease, box-shadow 0.15s ease;
}
.styles-module__webhookUrlInput___2375C::placeholder {
  color: rgba(255, 255, 255, 0.3);
}
.styles-module__webhookUrlInput___2375C:focus {
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.08);
}
[data-agentation-theme=light] .styles-module__webhookUrlInput___2375C {
  border-color: rgba(0, 0, 0, 0.1);
  background: rgba(0, 0, 0, 0.03);
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__webhookUrlInput___2375C::placeholder {
  color: rgba(0, 0, 0, 0.3);
}
[data-agentation-theme=light] .styles-module__webhookUrlInput___2375C:focus {
  border-color: rgba(0, 0, 0, 0.25);
  background: rgba(0, 0, 0, 0.05);
}

.styles-module__settingsHeader___pwDY9 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 24px;
  margin-bottom: 0.5rem;
  padding-bottom: 9px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
}

.styles-module__settingsBrand___0gJeM {
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: -0.0094em;
  color: #fff;
  text-decoration: none;
}

.styles-module__settingsBrandSlash___uTG18 {
  color: var(--agentation-color-accent);
  transition: color 0.2s ease;
}

.styles-module__settingsVersion___TUcFq {
  font-size: 11px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.4);
  margin-left: auto;
  letter-spacing: -0.0094em;
}

.styles-module__settingsSection___m-YM2 + .styles-module__settingsSection___m-YM2 {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
}
.styles-module__settingsSection___m-YM2.styles-module__settingsSectionExtraPadding___jdhFV {
  padding-top: calc(0.5rem + 4px);
}

.styles-module__settingsSectionGrow___h-5HZ {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.styles-module__settingsRow___3sdhc {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 24px;
}
.styles-module__settingsRow___3sdhc.styles-module__settingsRowMarginTop___zA0Sp {
  margin-top: 8px;
}

.styles-module__dropdownContainer___BVnxe {
  position: relative;
}

.styles-module__dropdownButton___16NPz {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0.5rem;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
  letter-spacing: -0.0094em;
}
.styles-module__dropdownButton___16NPz:hover {
  background: rgba(255, 255, 255, 0.08);
}
.styles-module__dropdownButton___16NPz svg {
  opacity: 0.6;
}

.styles-module__cycleButton___FMKfw {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0;
  border: none;
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #fff;
  cursor: pointer;
  letter-spacing: -0.0094em;
}
[data-agentation-theme=light] .styles-module__cycleButton___FMKfw {
  color: rgba(0, 0, 0, 0.85);
}
.styles-module__cycleButton___FMKfw:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.styles-module__settingsRowDisabled___EgS0V .styles-module__settingsLabel___8UjfX {
  color: rgba(255, 255, 255, 0.2);
}
[data-agentation-theme=light] .styles-module__settingsRowDisabled___EgS0V .styles-module__settingsLabel___8UjfX {
  color: rgba(0, 0, 0, 0.2);
}
.styles-module__settingsRowDisabled___EgS0V .styles-module__toggleSwitch___l4Ygm {
  opacity: 0.4;
  cursor: not-allowed;
}

@keyframes styles-module__cycleTextIn___Q6zJf {
  0% {
    opacity: 0;
    transform: translateY(-6px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}
.styles-module__cycleButtonText___fD1LR {
  display: inline-block;
  animation: styles-module__cycleTextIn___Q6zJf 0.2s ease-out;
}

.styles-module__cycleDots___LWuoQ {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.styles-module__cycleDot___nPgLY {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  transform: scale(0.667);
  transition: background-color 0.25s ease-out, transform 0.25s ease-out;
}
.styles-module__cycleDot___nPgLY.styles-module__active___-zoN6 {
  background: #fff;
  transform: scale(1);
}
[data-agentation-theme=light] .styles-module__cycleDot___nPgLY {
  background: rgba(0, 0, 0, 0.2);
}
[data-agentation-theme=light] .styles-module__cycleDot___nPgLY.styles-module__active___-zoN6 {
  background: rgba(0, 0, 0, 0.7);
}

.styles-module__dropdownMenu___k73ER {
  position: absolute;
  right: 0;
  top: calc(100% + 0.25rem);
  background: #1a1a1a;
  border-radius: 0.5rem;
  padding: 0.25rem;
  min-width: 120px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.1);
  z-index: 10;
  animation: styles-module__scaleIn___c-r1K 0.15s ease-out;
}

.styles-module__dropdownItem___ylsLj {
  width: 100%;
  display: flex;
  align-items: center;
  padding: 0.5rem 0.625rem;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  text-align: left;
  transition: background-color 0.15s ease, color 0.15s ease;
  letter-spacing: -0.0094em;
}
.styles-module__dropdownItem___ylsLj:hover {
  background: rgba(255, 255, 255, 0.08);
}
.styles-module__dropdownItem___ylsLj.styles-module__selected___OwRqP {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-weight: 600;
}

.styles-module__settingsLabel___8UjfX {
  font-size: 0.8125rem;
  font-weight: 400;
  letter-spacing: -0.0094em;
  color: rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  gap: 0.125rem;
}
[data-agentation-theme=light] .styles-module__settingsLabel___8UjfX {
  color: rgba(0, 0, 0, 0.5);
}

.styles-module__settingsLabelMarker___ewdtV {
  padding-top: 3px;
  margin-bottom: 10px;
}

.styles-module__settingsOptions___LyrBA {
  display: flex;
  gap: 0.25rem;
}

.styles-module__settingsOption___UNa12 {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.375rem 0.5rem;
  border: none;
  border-radius: 0.375rem;
  background: transparent;
  font-size: 0.6875rem;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.7);
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}
.styles-module__settingsOption___UNa12:hover {
  background: rgba(0, 0, 0, 0.05);
}
.styles-module__settingsOption___UNa12.styles-module__selected___OwRqP {
  background: color-mix(in srgb, var(--agentation-color-blue) 15%, transparent);
  color: var(--agentation-color-blue);
}

.styles-module__sliderContainer___ducXj {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.styles-module__slider___GLdxp {
  -webkit-appearance: none;
  appearance: none;
  width: 100%;
  height: 4px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}
.styles-module__slider___GLdxp::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}
.styles-module__slider___GLdxp::-moz-range-thumb {
  width: 14px;
  height: 14px;
  background: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}
.styles-module__slider___GLdxp:hover::-webkit-slider-thumb {
  transform: scale(1.15);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
}
.styles-module__slider___GLdxp:hover::-moz-range-thumb {
  transform: scale(1.15);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
}

.styles-module__sliderLabels___FhLDB {
  display: flex;
  justify-content: space-between;
}

.styles-module__sliderLabel___U8sPr {
  font-size: 0.625rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.4);
  cursor: pointer;
  transition: color 0.15s ease;
}
.styles-module__sliderLabel___U8sPr:hover {
  color: rgba(255, 255, 255, 0.7);
}
.styles-module__sliderLabel___U8sPr.styles-module__active___-zoN6 {
  color: rgba(255, 255, 255, 0.9);
}

.styles-module__colorOptions___iHCNX {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.375rem;
  margin-bottom: 1px;
}

.styles-module__colorOption___IodiY {
  display: block;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid transparent;
  background-color: var(--swatch);
  cursor: pointer;
  transition: transform 0.2s cubic-bezier(0.25, 1, 0.5, 1);
}
@supports (color: color(display-p3 0 0 0)) {
  .styles-module__colorOption___IodiY {
    background-color: var(--swatch-p3);
  }
}
.styles-module__colorOption___IodiY:hover {
  transform: scale(1.15);
}
.styles-module__colorOption___IodiY.styles-module__selected___OwRqP {
  transform: scale(0.83);
}

.styles-module__colorOptionRing___U2xpo {
  display: flex;
  width: 24px;
  height: 24px;
  border: 2px solid transparent;
  border-radius: 50%;
  transition: border-color 0.3s ease;
}
.styles-module__colorOptionRing___U2xpo.styles-module__selected___OwRqP {
  border-color: var(--swatch);
}
@supports (color: color(display-p3 0 0 0)) {
  .styles-module__colorOptionRing___U2xpo.styles-module__selected___OwRqP {
    border-color: var(--swatch-p3);
  }
}

.styles-module__settingsToggle___fBrFn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}
.styles-module__settingsToggle___fBrFn + .styles-module__settingsToggle___fBrFn {
  margin-top: calc(0.5rem + 6px);
}
.styles-module__settingsToggle___fBrFn input[type=checkbox] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.styles-module__settingsToggle___fBrFn.styles-module__settingsToggleMarginBottom___MZUyF {
  margin-bottom: calc(0.5rem + 6px);
}

@keyframes styles-module__mcpPulse___uNggr {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-green) 50%, transparent);
  }
  70% {
    box-shadow: 0 0 0 6px color-mix(in srgb, var(--agentation-color-green) 0%, transparent);
  }
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-green) 0%, transparent);
  }
}
@keyframes styles-module__mcpPulseError___fov9B {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-red) 50%, transparent);
  }
  70% {
    box-shadow: 0 0 0 6px color-mix(in srgb, var(--agentation-color-red) 0%, transparent);
  }
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-red) 0%, transparent);
  }
}
.styles-module__mcpStatusDot___ibgkc {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.styles-module__mcpStatusDot___ibgkc.styles-module__connecting___uo-CW {
  background-color: var(--agentation-color-yellow);
  animation: styles-module__mcpPulse___uNggr 1.5s infinite;
}
.styles-module__mcpStatusDot___ibgkc.styles-module__connected___7c28g {
  background-color: var(--agentation-color-green);
  animation: styles-module__mcpPulse___uNggr 2.5s ease-in-out infinite;
}
.styles-module__mcpStatusDot___ibgkc.styles-module__disconnected___cHPxR {
  background-color: var(--agentation-color-red);
  animation: styles-module__mcpPulseError___fov9B 2s infinite;
}

.styles-module__drawCanvas___7cG9U {
  position: fixed;
  inset: 0;
  z-index: 99996;
  pointer-events: none !important;
}
.styles-module__drawCanvas___7cG9U.styles-module__active___-zoN6 {
  pointer-events: auto !important;
  cursor: crosshair !important;
}
.styles-module__drawCanvas___7cG9U.styles-module__active___-zoN6[data-stroke-hover] {
  cursor: pointer !important;
}

.styles-module__dragSelection___kZLq2 {
  position: fixed;
  top: 0;
  left: 0;
  border: 2px solid color-mix(in srgb, var(--agentation-color-green) 60%, transparent);
  border-radius: 4px;
  background-color: color-mix(in srgb, var(--agentation-color-green) 8%, transparent);
  pointer-events: none;
  z-index: 99997;
  will-change: transform, width, height;
  contain: layout style;
}

.styles-module__dragCount___KM90j {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background-color: var(--agentation-color-green);
  color: white;
  font-size: 0.875rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  border-radius: 1rem;
  min-width: 1.5rem;
  text-align: center;
}

.styles-module__highlightsContainer___-0xzG {
  position: fixed;
  top: 0;
  left: 0;
  pointer-events: none;
  z-index: 99996;
}

.styles-module__selectedElementHighlight___fyVlI {
  position: fixed;
  top: 0;
  left: 0;
  border: 2px solid color-mix(in srgb, var(--agentation-color-green) 50%, transparent);
  border-radius: 4px;
  background: color-mix(in srgb, var(--agentation-color-green) 6%, transparent);
  pointer-events: none;
  will-change: transform, width, height;
  contain: layout style;
}

[data-agentation-theme=light] .styles-module__toolbarContainer___dIhma {
  background: #fff;
  color: rgba(0, 0, 0, 0.85);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08), 0 4px 16px rgba(0, 0, 0, 0.06), 0 0 0 1px rgba(0, 0, 0, 0.04);
}
[data-agentation-theme=light] .styles-module__toolbarContainer___dIhma.styles-module__collapsed___Rydsn:hover {
  background: #f5f5f5;
}
[data-agentation-theme=light] .styles-module__toggleContent___0yfyP {
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN {
  color: rgba(0, 0, 0, 0.5);
}
[data-agentation-theme=light] .styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN:hover {
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__toggleContent___0yfyP.styles-module__expandedToggle___F7SRN:hover::before {
  background: rgba(0, 0, 0, 0.06);
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc {
  color: rgba(0, 0, 0, 0.5);
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc:hover:not(:disabled):not([data-active=true]):not([data-failed=true]):not([data-auto-sync=true]):not([data-error=true]):not([data-no-hover=true]) {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc[data-active=true] {
  color: var(--agentation-color-blue);
  background: color-mix(in srgb, var(--agentation-color-blue) 15%, transparent);
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc[data-error=true] {
  color: var(--agentation-color-red);
  background: color-mix(in srgb, var(--agentation-color-red) 15%, transparent);
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc[data-danger]:hover:not(:disabled):not([data-active=true]):not([data-failed=true]) {
  color: var(--agentation-color-red);
  background: color-mix(in srgb, var(--agentation-color-red) 15%, transparent);
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc[data-auto-sync=true] {
  color: var(--agentation-color-green);
  background: transparent;
}
[data-agentation-theme=light] .styles-module__controlButton___8Q0jc[data-failed=true] {
  color: var(--agentation-color-red);
  background: color-mix(in srgb, var(--agentation-color-red) 15%, transparent);
}
[data-agentation-theme=light] .styles-module__buttonTooltip___Burd9 {
  background: #fff;
  color: rgba(0, 0, 0, 0.85);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08), 0 4px 16px rgba(0, 0, 0, 0.06), 0 0 0 1px rgba(0, 0, 0, 0.04);
}
[data-agentation-theme=light] .styles-module__buttonTooltip___Burd9::after {
  background: #fff;
}
[data-agentation-theme=light] .styles-module__divider___c--s1 {
  background: rgba(0, 0, 0, 0.1);
}`,le={toolbar:"styles-module__toolbar___wNsdK",markersLayer:"styles-module__markersLayer___-25j1",fixedMarkersLayer:"styles-module__fixedMarkersLayer___ffyX6",controlsContent:"styles-module__controlsContent___9GJWU",disableTransitions:"styles-module__disableTransitions___EopxO",positionContext:"styles-module__positionContext___AZFHE",toolbarContainer:"styles-module__toolbarContainer___dIhma",entrance:"styles-module__entrance___sgHd8",toolbarEnter:"styles-module__toolbarEnter___u8RRu",hiding:"styles-module__hiding___1td44",toolbarHide:"styles-module__toolbarHide___y8kaT",collapsed:"styles-module__collapsed___Rydsn",expanded:"styles-module__expanded___ofKPx",serverConnected:"styles-module__serverConnected___Gfbou",buttonWrapper:"styles-module__buttonWrapper___rBcdv",toggleWrapper:"styles-module__toggleWrapper___7N0-q",togglePlaceholder:"styles-module__togglePlaceholder___wnqrL",toggleContent:"styles-module__toggleContent___0yfyP",expandedToggle:"styles-module__expandedToggle___F7SRN",toggleGlyph:"styles-module__toggleGlyph___R7Oom",toggleIcon:"styles-module__toggleIcon___Jbtus",toggleTopLine:"styles-module__toggleTopLine___hQaCm",toggleMiddleLine:"styles-module__toggleMiddleLine___sFFVe",toggleBottomLine:"styles-module__toggleBottomLine___V-jX3",toggleSparkle:"styles-module__toggleSparkle___eeF99",controlButton:"styles-module__controlButton___8Q0jc",visible:"styles-module__visible___KHwEW",hidden:"styles-module__hidden___Ae8H4",badge:"styles-module__badge___2XsgF",fadeOut:"styles-module__fadeOut___6Ut6-",badgeEnter:"styles-module__badgeEnter___mVQLj",statusShowing:"styles-module__statusShowing___te6iu",buttonBadge:"styles-module__buttonBadge___NeFWb",mcpIndicator:"styles-module__mcpIndicator___zGJeL",connected:"styles-module__connected___7c28g",mcpIndicatorPulseConnected:"styles-module__mcpIndicatorPulseConnected___EDodZ",connecting:"styles-module__connecting___uo-CW",mcpIndicatorPulseConnecting:"styles-module__mcpIndicatorPulseConnecting___cCYte",connectionIndicatorWrapper:"styles-module__connectionIndicatorWrapper___L-e-3",connectionIndicator:"styles-module__connectionIndicator___afk9p",connectionIndicatorVisible:"styles-module__connectionIndicatorVisible___C-i5B",connectionIndicatorConnected:"styles-module__connectionIndicatorConnected___IY8pR",connectionPulse:"styles-module__connectionPulse___-Zycw",connectionIndicatorDisconnected:"styles-module__connectionIndicatorDisconnected___kmpaZ",connectionIndicatorConnecting:"styles-module__connectionIndicatorConnecting___QmSLH",buttonTooltip:"styles-module__buttonTooltip___Burd9",tooltipsInSession:"styles-module__tooltipsInSession___-0lHH",sendButtonWrapper:"styles-module__sendButtonWrapper___UUxG6",sendButtonVisible:"styles-module__sendButtonVisible___WPSQU",shortcut:"styles-module__shortcut___lEAQk",tooltipBelow:"styles-module__tooltipBelow___m6ats",tooltipsHidden:"styles-module__tooltipsHidden___VtLJG",tooltipVisible:"styles-module__tooltipVisible___0jcCv",buttonWrapperAlignLeft:"styles-module__buttonWrapperAlignLeft___myzIp",buttonWrapperAlignRight:"styles-module__buttonWrapperAlignRight___HCQFR",divider:"styles-module__divider___c--s1",overlay:"styles-module__overlay___Q1O9y",hoverHighlight:"styles-module__hoverHighlight___ogakW",enter:"styles-module__enter___WFIki",hoverHighlightIn:"styles-module__hoverHighlightIn___6WYHY",multiSelectOutline:"styles-module__multiSelectOutline___cSJ-m",fadeIn:"styles-module__fadeIn___b9qmf",exit:"styles-module__exit___fyOJ0",singleSelectOutline:"styles-module__singleSelectOutline___QhX-O",hoverTooltip:"styles-module__hoverTooltip___bvLk7",hoverTooltipIn:"styles-module__hoverTooltipIn___FYGQx",hoverReactPath:"styles-module__hoverReactPath___gx1IJ",hoverElementName:"styles-module__hoverElementName___QMLMl",marker:"styles-module__marker___6sQrs",clearing:"styles-module__clearing___FQ--7",markerIn:"styles-module__markerIn___5FaAP",markerOut:"styles-module__markerOut___GU5jX",pending:"styles-module__pending___2IHLC",fixed:"styles-module__fixed___dBMHC",multiSelect:"styles-module__multiSelect___YWiuz",hovered:"styles-module__hovered___ZgXIy",renumber:"styles-module__renumber___nCTxD",renumberRoll:"styles-module__renumberRoll___Wgbq3",markerTooltip:"styles-module__markerTooltip___aLJID",tooltipIn:"styles-module__tooltipIn___0N31w",markerQuote:"styles-module__markerQuote___FHmrz",markerNote:"styles-module__markerNote___QkrrS",markerHint:"styles-module__markerHint___2iF-6",settingsPanel:"styles-module__settingsPanel___OxX3Y",settingsHeader:"styles-module__settingsHeader___pwDY9",settingsBrand:"styles-module__settingsBrand___0gJeM",settingsBrandSlash:"styles-module__settingsBrandSlash___uTG18",settingsVersion:"styles-module__settingsVersion___TUcFq",settingsSection:"styles-module__settingsSection___m-YM2",settingsLabel:"styles-module__settingsLabel___8UjfX",cycleButton:"styles-module__cycleButton___FMKfw",cycleDot:"styles-module__cycleDot___nPgLY",dropdownButton:"styles-module__dropdownButton___16NPz",toggleLabel:"styles-module__toggleLabel___Xm8Aa",customCheckbox:"styles-module__customCheckbox___U39ax",sliderLabel:"styles-module__sliderLabel___U8sPr",slider:"styles-module__slider___GLdxp",themeToggle:"styles-module__themeToggle___2rUjA",settingsOption:"styles-module__settingsOption___UNa12",selected:"styles-module__selected___OwRqP",settingsPanelContainer:"styles-module__settingsPanelContainer___Xksv8",settingsPage:"styles-module__settingsPage___6YfHH",slideLeft:"styles-module__slideLeft___Ps01J",automationsPage:"styles-module__automationsPage___uvCq6",slideIn:"styles-module__slideIn___4-qXe",settingsNavLink:"styles-module__settingsNavLink___wCzJt",settingsNavLinkRight:"styles-module__settingsNavLinkRight___ZWwhj",mcpNavIndicator:"styles-module__mcpNavIndicator___cl9pO",mcpPulse:"styles-module__mcpPulse___uNggr",settingsBackButton:"styles-module__settingsBackButton___bIe2j",automationHeader:"styles-module__automationHeader___InP0r",automationDescription:"styles-module__automationDescription___NKlmo",learnMoreLink:"styles-module__learnMoreLink___8xv-x",autoSendRow:"styles-module__autoSendRow___UblX5",autoSendLabel:"styles-module__autoSendLabel___icDc2",active:"styles-module__active___-zoN6",webhookUrlInput:"styles-module__webhookUrlInput___2375C",settingsSectionExtraPadding:"styles-module__settingsSectionExtraPadding___jdhFV",settingsSectionGrow:"styles-module__settingsSectionGrow___h-5HZ",settingsRow:"styles-module__settingsRow___3sdhc",settingsRowMarginTop:"styles-module__settingsRowMarginTop___zA0Sp",dropdownContainer:"styles-module__dropdownContainer___BVnxe",settingsRowDisabled:"styles-module__settingsRowDisabled___EgS0V",toggleSwitch:"styles-module__toggleSwitch___l4Ygm",cycleButtonText:"styles-module__cycleButtonText___fD1LR",cycleTextIn:"styles-module__cycleTextIn___Q6zJf",cycleDots:"styles-module__cycleDots___LWuoQ",dropdownMenu:"styles-module__dropdownMenu___k73ER",scaleIn:"styles-module__scaleIn___c-r1K",dropdownItem:"styles-module__dropdownItem___ylsLj",settingsLabelMarker:"styles-module__settingsLabelMarker___ewdtV",settingsOptions:"styles-module__settingsOptions___LyrBA",sliderContainer:"styles-module__sliderContainer___ducXj",sliderLabels:"styles-module__sliderLabels___FhLDB",colorOptions:"styles-module__colorOptions___iHCNX",colorOption:"styles-module__colorOption___IodiY",colorOptionRing:"styles-module__colorOptionRing___U2xpo",settingsToggle:"styles-module__settingsToggle___fBrFn",settingsToggleMarginBottom:"styles-module__settingsToggleMarginBottom___MZUyF",mcpStatusDot:"styles-module__mcpStatusDot___ibgkc",disconnected:"styles-module__disconnected___cHPxR",mcpPulseError:"styles-module__mcpPulseError___fov9B",drawCanvas:"styles-module__drawCanvas___7cG9U",dragSelection:"styles-module__dragSelection___kZLq2",dragCount:"styles-module__dragCount___KM90j",highlightsContainer:"styles-module__highlightsContainer___-0xzG",selectedElementHighlight:"styles-module__selectedElementHighlight___fyVlI",scaleOut:"styles-module__scaleOut___Wctwz",slideUp:"styles-module__slideUp___kgD36",slideDown:"styles-module__slideDown___zcdje"};function Rv({active:s}){return o.jsxs("svg",{className:le.toggleGlyph,"data-active":s,width:"24",height:"24",viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round","aria-hidden":"true",children:[o.jsx("path",{className:le.toggleTopLine,d:"M5.5 6.75H18.5"}),o.jsx("path",{className:le.toggleMiddleLine,d:"M5.5 12H11.5"}),o.jsx("path",{className:le.toggleBottomLine,d:"M5.5 17.25H9.25"}),o.jsx("path",{className:le.toggleSparkle,d:"M16 12.75L16.5179 13.9677C16.8078 14.6494 17.3506 15.1922 18.0323 15.4821L19.25 16L18.0323 16.5179C17.3506 16.8078 16.8078 17.3506 16.5179 18.0323L16 19.25L15.4821 18.0323C15.1922 17.3506 14.6494 16.8078 13.9677 16.5179L12.75 16L13.9677 15.4821C14.6494 15.1922 15.1922 14.6494 15.4821 13.9677L16 12.75Z"})]})}var Ae={navigation:{width:800,height:56},hero:{width:800,height:320},header:{width:800,height:80},section:{width:800,height:400},sidebar:{width:240,height:400},footer:{width:800,height:160},modal:{width:480,height:300},card:{width:280,height:240},text:{width:400,height:120},image:{width:320,height:200},video:{width:480,height:270},table:{width:560,height:220},grid:{width:600,height:300},list:{width:300,height:180},chart:{width:400,height:240},button:{width:140,height:40},input:{width:280,height:56},form:{width:360,height:320},tabs:{width:480,height:240},dropdown:{width:200,height:200},toggle:{width:44,height:24},search:{width:320,height:44},avatar:{width:48,height:48},badge:{width:80,height:28},breadcrumb:{width:300,height:24},pagination:{width:300,height:36},progress:{width:240,height:8},divider:{width:600,height:1},accordion:{width:400,height:200},carousel:{width:600,height:300},toast:{width:320,height:64},tooltip:{width:180,height:40},pricing:{width:300,height:360},testimonial:{width:360,height:200},cta:{width:600,height:160},alert:{width:400,height:56},banner:{width:800,height:48},stat:{width:200,height:120},stepper:{width:480,height:48},tag:{width:72,height:28},rating:{width:160,height:28},map:{width:480,height:300},timeline:{width:360,height:320},fileUpload:{width:360,height:180},codeBlock:{width:480,height:200},calendar:{width:300,height:300},notification:{width:360,height:72},productCard:{width:280,height:360},profile:{width:280,height:200},drawer:{width:320,height:400},popover:{width:240,height:160},logo:{width:120,height:40},faq:{width:560,height:320},gallery:{width:560,height:360},checkbox:{width:20,height:20},radio:{width:20,height:20},slider:{width:240,height:32},datePicker:{width:300,height:320},skeleton:{width:320,height:120},chip:{width:96,height:32},icon:{width:24,height:24},spinner:{width:32,height:32},feature:{width:360,height:200},team:{width:560,height:280},login:{width:360,height:360},contact:{width:400,height:320}},ey=[{section:"Layout",items:[{type:"navigation",label:"Navigation",...Ae.navigation},{type:"header",label:"Header",...Ae.header},{type:"hero",label:"Hero",...Ae.hero},{type:"section",label:"Section",...Ae.section},{type:"sidebar",label:"Sidebar",...Ae.sidebar},{type:"footer",label:"Footer",...Ae.footer},{type:"modal",label:"Modal",...Ae.modal},{type:"banner",label:"Banner",...Ae.banner},{type:"drawer",label:"Drawer",...Ae.drawer},{type:"popover",label:"Popover",...Ae.popover},{type:"divider",label:"Divider",...Ae.divider}]},{section:"Content",items:[{type:"card",label:"Card",...Ae.card},{type:"text",label:"Text",...Ae.text},{type:"image",label:"Image",...Ae.image},{type:"video",label:"Video",...Ae.video},{type:"table",label:"Table",...Ae.table},{type:"grid",label:"Grid",...Ae.grid},{type:"list",label:"List",...Ae.list},{type:"chart",label:"Chart",...Ae.chart},{type:"codeBlock",label:"Code Block",...Ae.codeBlock},{type:"map",label:"Map",...Ae.map},{type:"timeline",label:"Timeline",...Ae.timeline},{type:"calendar",label:"Calendar",...Ae.calendar},{type:"accordion",label:"Accordion",...Ae.accordion},{type:"carousel",label:"Carousel",...Ae.carousel},{type:"logo",label:"Logo",...Ae.logo},{type:"faq",label:"FAQ",...Ae.faq},{type:"gallery",label:"Gallery",...Ae.gallery}]},{section:"Controls",items:[{type:"button",label:"Button",...Ae.button},{type:"input",label:"Input",...Ae.input},{type:"search",label:"Search",...Ae.search},{type:"form",label:"Form",...Ae.form},{type:"tabs",label:"Tabs",...Ae.tabs},{type:"dropdown",label:"Dropdown",...Ae.dropdown},{type:"toggle",label:"Toggle",...Ae.toggle},{type:"stepper",label:"Stepper",...Ae.stepper},{type:"rating",label:"Rating",...Ae.rating},{type:"fileUpload",label:"File Upload",...Ae.fileUpload},{type:"checkbox",label:"Checkbox",...Ae.checkbox},{type:"radio",label:"Radio",...Ae.radio},{type:"slider",label:"Slider",...Ae.slider},{type:"datePicker",label:"Date Picker",...Ae.datePicker}]},{section:"Elements",items:[{type:"avatar",label:"Avatar",...Ae.avatar},{type:"badge",label:"Badge",...Ae.badge},{type:"tag",label:"Tag",...Ae.tag},{type:"breadcrumb",label:"Breadcrumb",...Ae.breadcrumb},{type:"pagination",label:"Pagination",...Ae.pagination},{type:"progress",label:"Progress",...Ae.progress},{type:"alert",label:"Alert",...Ae.alert},{type:"toast",label:"Toast",...Ae.toast},{type:"notification",label:"Notification",...Ae.notification},{type:"tooltip",label:"Tooltip",...Ae.tooltip},{type:"stat",label:"Stat",...Ae.stat},{type:"skeleton",label:"Skeleton",...Ae.skeleton},{type:"chip",label:"Chip",...Ae.chip},{type:"icon",label:"Icon",...Ae.icon},{type:"spinner",label:"Spinner",...Ae.spinner}]},{section:"Blocks",items:[{type:"pricing",label:"Pricing",...Ae.pricing},{type:"testimonial",label:"Testimonial",...Ae.testimonial},{type:"cta",label:"CTA",...Ae.cta},{type:"productCard",label:"Product Card",...Ae.productCard},{type:"profile",label:"Profile",...Ae.profile},{type:"feature",label:"Feature",...Ae.feature},{type:"team",label:"Team",...Ae.team},{type:"login",label:"Login",...Ae.login},{type:"contact",label:"Contact",...Ae.contact}]}],mo={};for(const s of ey)for(const r of s.items)mo[r.type]=r;function se({w:s,h:r=3,strong:c}){return o.jsx("div",{style:{width:typeof s=="number"?`${s}px`:s,height:r,borderRadius:2,background:c?"var(--agd-bar-strong)":"var(--agd-bar)",flexShrink:0}})}function en({w:s,h:r,radius:c=3,style:u}){return o.jsx("div",{style:{width:typeof s=="number"?`${s}px`:s,height:typeof r=="number"?`${r}px`:r,borderRadius:c,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",flexShrink:0,...u}})}function bl({size:s}){return o.jsx("div",{style:{width:s,height:s,borderRadius:"50%",border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",flexShrink:0}})}function zv({width:s,height:r}){const c=Math.max(8,r*.2);return o.jsxs("div",{style:{display:"flex",alignItems:"center",height:"100%",padding:`0 ${c}px`,gap:s*.02},children:[o.jsx(en,{w:Math.max(20,r*.5),h:Math.max(12,r*.4),radius:2}),o.jsxs("div",{style:{flex:1,display:"flex",gap:s*.03,marginLeft:s*.04},children:[o.jsx(se,{w:s*.06}),o.jsx(se,{w:s*.07}),o.jsx(se,{w:s*.05}),o.jsx(se,{w:s*.06})]}),o.jsx(en,{w:s*.1,h:Math.min(28,r*.5),radius:4})]})}function Av({width:s,height:r,text:c}){return o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",height:"100%",gap:r*.05},children:[c?o.jsx("span",{style:{fontSize:Math.min(20,r*.08),fontWeight:600,color:"var(--agd-text-3)",textAlign:"center",maxWidth:"80%"},children:c}):o.jsx(se,{w:s*.5,h:Math.max(6,r*.04),strong:!0}),o.jsx(se,{w:s*.6}),o.jsx(se,{w:s*.4}),o.jsx(en,{w:Math.min(140,s*.2),h:Math.min(36,r*.12),radius:6,style:{marginTop:r*.06}})]})}function Ov({width:s,height:r}){const c=Math.max(3,Math.floor(r/36));return o.jsxs("div",{style:{padding:s*.08,display:"flex",flexDirection:"column",gap:r*.03},children:[o.jsx(se,{w:s*.6,h:4,strong:!0}),Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:6},children:[o.jsx(en,{w:10,h:10,radius:2}),o.jsx(se,{w:s*(.4+f*17%30/100)})]},f))]})}function Dv({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(s/160)));return o.jsx("div",{style:{display:"flex",padding:`${r*.12}px ${s*.03}px`,gap:s*.05},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:4},children:[o.jsx(se,{w:"60%",h:3,strong:!0}),o.jsx(se,{w:"80%",h:2}),o.jsx(se,{w:"70%",h:2}),o.jsx(se,{w:"60%",h:2})]},f))})}function Lv({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsxs("div",{style:{padding:"10px 12px",borderBottom:"1px solid var(--agd-stroke)",display:"flex",alignItems:"center",justifyContent:"space-between"},children:[o.jsx(se,{w:s*.3,h:4,strong:!0}),o.jsx("div",{style:{width:14,height:14,border:"1px solid var(--agd-stroke)",borderRadius:3}})]}),o.jsxs("div",{style:{flex:1,padding:12,display:"flex",flexDirection:"column",gap:6},children:[o.jsx(se,{w:"90%"}),o.jsx(se,{w:"70%"}),o.jsx(se,{w:"80%"})]}),o.jsxs("div",{style:{padding:"10px 12px",borderTop:"1px solid var(--agd-stroke)",display:"flex",justifyContent:"flex-end",gap:8},children:[o.jsx(en,{w:70,h:26,radius:4}),o.jsx(en,{w:70,h:26,radius:4,style:{background:"var(--agd-bar)"}})]})]})}function Bv({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsx("div",{style:{height:"40%",background:"var(--agd-fill)",borderBottom:"1px dashed var(--agd-stroke)"}}),o.jsxs("div",{style:{flex:1,padding:10,display:"flex",flexDirection:"column",gap:5},children:[o.jsx(se,{w:"70%",h:4,strong:!0}),o.jsx(se,{w:"95%",h:2}),o.jsx(se,{w:"85%",h:2}),o.jsx(se,{w:"50%",h:2})]})]})}function $v({width:s,height:r,text:c}){if(c)return o.jsx("div",{style:{padding:4,fontSize:Math.min(14,r*.3),lineHeight:1.5,color:"var(--agd-text-3)",wordBreak:"break-word",overflow:"hidden"},children:c});const u=Math.max(2,Math.floor(r/18));return o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:6,padding:4},children:[o.jsx(se,{w:s*.6,h:5,strong:!0}),Array.from({length:u},(f,d)=>o.jsx(se,{w:`${70+d*13%25}%`,h:2},d))]})}function Uv({width:s,height:r}){return o.jsx("div",{style:{height:"100%",position:"relative"},children:o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,preserveAspectRatio:"none",fill:"none",children:[o.jsx("line",{x1:"0",y1:"0",x2:s,y2:r,stroke:"var(--agd-stroke)",strokeWidth:"1"}),o.jsx("line",{x1:s,y1:"0",x2:"0",y2:r,stroke:"var(--agd-stroke)",strokeWidth:"1"}),o.jsx("circle",{cx:s*.3,cy:r*.3,r:Math.min(s,r)*.08,fill:"var(--agd-fill)",stroke:"var(--agd-stroke)",strokeWidth:"0.8"})]})})}function Hv({width:s,height:r}){const c=Math.max(2,Math.min(5,Math.floor(s/100))),u=Math.max(2,Math.min(6,Math.floor(r/32)));return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsx("div",{style:{display:"flex",borderBottom:"1px solid var(--agd-stroke)",padding:"6px 0"},children:Array.from({length:c},(f,d)=>o.jsx("div",{style:{flex:1,padding:"0 8px"},children:o.jsx(se,{w:"70%",h:3,strong:!0})},d))}),Array.from({length:u},(f,d)=>o.jsx("div",{style:{display:"flex",borderBottom:"1px solid rgba(255,255,255,0.03)",padding:"6px 0"},children:Array.from({length:c},(m,g)=>o.jsx("div",{style:{flex:1,padding:"0 8px"},children:o.jsx(se,{w:`${50+(d*7+g*13)%40}%`,h:2})},g))},d))]})}function Yv({width:s,height:r}){const c=Math.max(2,Math.floor(r/28));return o.jsx("div",{style:{display:"flex",flexDirection:"column",gap:4,padding:4},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:8,padding:"4px 0"},children:[o.jsx(bl,{size:8}),o.jsx(se,{w:`${55+f*17%35}%`,h:2})]},f))})}function qv({width:s,height:r,text:c}){return o.jsx("div",{style:{height:"100%",borderRadius:Math.min(8,r/3),border:"1px solid var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",justifyContent:"center"},children:c?o.jsx("span",{style:{fontSize:Math.min(13,r*.4),fontWeight:500,color:"var(--agd-text-3)",letterSpacing:"-0.01em"},children:c}):o.jsx(se,{w:Math.max(20,s*.5),h:3,strong:!0})})}function Iv({width:s,height:r}){return o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:4,height:"100%",justifyContent:"center"},children:[o.jsx(se,{w:Math.min(80,s*.3),h:2}),o.jsx("div",{style:{height:Math.min(36,r*.6),borderRadius:4,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",paddingLeft:8},children:o.jsx(se,{w:"40%",h:2})})]})}function Xv({width:s,height:r}){const c=Math.max(2,Math.min(5,Math.floor(r/56)));return o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:r*.04,padding:8},children:[Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:4},children:[o.jsx(se,{w:60+f*17%30,h:2}),o.jsx(en,{w:"100%",h:28,radius:4})]},f)),o.jsx(en,{w:Math.min(120,s*.35),h:30,radius:6,style:{marginTop:8,alignSelf:"flex-end",background:"var(--agd-bar)"}})]})}function Qv({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(s/120)));return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsx("div",{style:{display:"flex",gap:2,borderBottom:"1px solid var(--agd-stroke)"},children:Array.from({length:c},(u,f)=>o.jsx("div",{style:{padding:"8px 12px",borderBottom:f===0?"2px solid var(--agd-bar-strong)":"none"},children:o.jsx(se,{w:60,h:3,strong:f===0})},f))}),o.jsxs("div",{style:{flex:1,padding:12,display:"flex",flexDirection:"column",gap:6},children:[o.jsx(se,{w:"80%",h:2}),o.jsx(se,{w:"65%",h:2}),o.jsx(se,{w:"75%",h:2})]})]})}function Gv({width:s,height:r}){const c=Math.min(s,r)/2;return o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:[o.jsx("circle",{cx:s/2,cy:r/2,r:c-1,stroke:"var(--agd-stroke)",fill:"var(--agd-fill)",strokeWidth:"1.5",strokeDasharray:"3 2"}),o.jsx("circle",{cx:s/2,cy:r*.38,r:c*.28,stroke:"var(--agd-stroke)",fill:"var(--agd-fill)",strokeWidth:"0.8"}),o.jsx("path",{d:`M${s/2-c*.55} ${r*.78} C${s/2-c*.55} ${r*.55} ${s/2+c*.55} ${r*.55} ${s/2+c*.55} ${r*.78}`,stroke:"var(--agd-stroke)",fill:"var(--agd-fill)",strokeWidth:"0.8"})]})}function Vv({width:s,height:r}){return o.jsx("div",{style:{height:"100%",borderRadius:r/2,border:"1px solid var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",justifyContent:"center"},children:o.jsx(se,{w:Math.max(16,s*.5),h:2,strong:!0})})}function Fv({width:s,height:r}){return o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",height:"100%",gap:r*.08},children:[o.jsx(se,{w:s*.5,h:Math.max(5,r*.06),strong:!0}),o.jsx(se,{w:s*.35})]})}function Wv({width:s,height:r}){return o.jsxs("div",{style:{display:"flex",flexDirection:"column",height:"100%",gap:r*.04,padding:s*.04},children:[o.jsx(se,{w:s*.3,h:4,strong:!0}),o.jsx(se,{w:s*.7}),o.jsx(se,{w:s*.5}),o.jsxs("div",{style:{flex:1,display:"flex",gap:s*.03,marginTop:r*.06},children:[o.jsx(en,{w:"33%",h:"100%",radius:4}),o.jsx(en,{w:"33%",h:"100%",radius:4}),o.jsx(en,{w:"33%",h:"100%",radius:4})]})]})}function Kv({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(s/140))),u=Math.max(1,Math.min(3,Math.floor(r/120)));return o.jsx("div",{style:{display:"grid",gridTemplateColumns:`repeat(${c}, 1fr)`,gridTemplateRows:`repeat(${u}, 1fr)`,gap:6,height:"100%"},children:Array.from({length:c*u},(f,d)=>o.jsx(en,{w:"100%",h:"100%",radius:4},d))})}function Zv({width:s,height:r}){const c=Math.max(2,Math.floor((r-32)/28));return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsx("div",{style:{padding:"6px 8px",borderBottom:"1px solid var(--agd-stroke)"},children:o.jsx(se,{w:s*.5,h:3,strong:!0})}),o.jsx("div",{style:{flex:1,padding:4,display:"flex",flexDirection:"column",gap:2},children:Array.from({length:c},(u,f)=>o.jsx("div",{style:{padding:"4px 6px",borderRadius:3,background:f===0?"var(--agd-fill)":"transparent"},children:o.jsx(se,{w:`${50+f*17%35}%`,h:2,strong:f===0})},f))})]})}function Jv({width:s,height:r}){const c=Math.min(s,r)/2;return o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:[o.jsx("rect",{x:"1",y:"1",width:s-2,height:r-2,rx:c,stroke:"var(--agd-stroke)",strokeWidth:"1"}),o.jsx("circle",{cx:s-c,cy:r/2,r:c*.7,fill:"var(--agd-bar)"})]})}function Pv({width:s,height:r}){const c=Math.min(r/2,20);return o.jsxs("div",{style:{height:"100%",borderRadius:c,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",padding:`0 ${c*.6}px`,gap:6},children:[o.jsx(bl,{size:Math.min(14,r*.4)}),o.jsx(se,{w:"50%",h:2})]})}function e5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",borderRadius:8,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",padding:"0 10px",gap:8},children:[o.jsx(bl,{size:Math.min(20,r*.5)}),o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:"60%",h:3,strong:!0}),o.jsx(se,{w:"80%",h:2})]}),o.jsx("div",{style:{width:14,height:14,border:"1px solid var(--agd-stroke)",borderRadius:3,flexShrink:0}})]})}function t5({width:s,height:r}){return o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:[o.jsx("rect",{x:"0",y:"0",width:s,height:r,rx:r/2,stroke:"var(--agd-stroke)",strokeWidth:"0.8"}),o.jsx("rect",{x:"1",y:"1",width:s*.65,height:r-2,rx:(r-2)/2,fill:"var(--agd-bar)"})]})}function n5({width:s,height:r}){const c=Math.max(3,Math.min(7,Math.floor(s/50))),u=s/(c*2);return o.jsx("div",{style:{height:"100%",display:"flex",alignItems:"flex-end",justifyContent:"space-around",padding:"0 4px",borderBottom:"1px solid var(--agd-stroke)"},children:Array.from({length:c},(f,d)=>{const m=30+(d*37+17)%55;return o.jsx(en,{w:u,h:`${m}%`,radius:2},d)})})}function l5({width:s,height:r}){const c=Math.min(s,r)*.12;return o.jsxs("div",{style:{height:"100%",position:"relative",display:"flex",alignItems:"center",justifyContent:"center"},children:[o.jsx(en,{w:"100%",h:"100%",radius:4}),o.jsx("div",{style:{position:"absolute",width:c*2,height:c*2,borderRadius:"50%",border:"1.5px solid var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",justifyContent:"center"},children:o.jsx("div",{style:{width:0,height:0,borderLeft:`${c*.6}px solid var(--agd-bar-strong)`,borderTop:`${c*.4}px solid transparent`,borderBottom:`${c*.4}px solid transparent`,marginLeft:c*.15}})})]})}function o5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center"},children:[o.jsx("div",{style:{flex:1,width:"100%",borderRadius:6,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",justifyContent:"center"},children:o.jsx(se,{w:"60%",h:2})}),o.jsx("div",{style:{width:8,height:8,background:"var(--agd-fill)",border:"1px dashed var(--agd-stroke)",borderTop:"none",borderLeft:"none",transform:"rotate(45deg)",marginTop:-5}})]})}function s5({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(s/80)));return o.jsx("div",{style:{display:"flex",alignItems:"center",height:"100%",gap:4},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:4},children:[f>0&&o.jsx("span",{style:{color:"var(--agd-stroke)",fontSize:10},children:"/"}),o.jsx(se,{w:40+f*13%20,h:2,strong:f===c-1})]},f))})}function a5({width:s,height:r}){const c=Math.max(3,Math.min(5,Math.floor(s/40))),u=Math.min(28,r*.8);return o.jsx("div",{style:{display:"flex",alignItems:"center",justifyContent:"center",height:"100%",gap:4},children:Array.from({length:c},(f,d)=>o.jsx(en,{w:u,h:u,radius:4,style:d===1?{background:"var(--agd-bar)"}:void 0},d))})}function r5({width:s}){return o.jsx("div",{style:{display:"flex",alignItems:"center",height:"100%"},children:o.jsx("div",{style:{width:"100%",height:1,background:"var(--agd-stroke)"}})})}function i5({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(r/40)));return o.jsx("div",{style:{display:"flex",flexDirection:"column",height:"100%"},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{borderBottom:"1px solid var(--agd-stroke)",padding:"8px 6px",display:"flex",alignItems:"center",justifyContent:"space-between",flex:f===0?2:1},children:[o.jsx(se,{w:`${40+f*17%25}%`,h:3,strong:!0}),o.jsx("span",{style:{fontSize:8,color:"var(--agd-stroke)"},children:f===0?"▼":"▶"})]},f))})}function c5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",gap:6},children:[o.jsxs("div",{style:{flex:1,display:"flex",gap:6,alignItems:"center"},children:[o.jsx("span",{style:{fontSize:12,color:"var(--agd-stroke)"},children:"‹"}),o.jsx(en,{w:"100%",h:"100%",radius:4}),o.jsx("span",{style:{fontSize:12,color:"var(--agd-stroke)"},children:"›"})]}),o.jsxs("div",{style:{display:"flex",justifyContent:"center",gap:4},children:[o.jsx(bl,{size:5}),o.jsx(bl,{size:5}),o.jsx(bl,{size:5})]})]})}function u5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center",padding:10,gap:r*.04},children:[o.jsx(se,{w:s*.4,h:3,strong:!0}),o.jsx(se,{w:s*.3,h:6,strong:!0}),o.jsx("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:4,width:"100%",padding:"8px 0"},children:Array.from({length:4},(c,u)=>o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:4},children:[o.jsx(bl,{size:5}),o.jsx(se,{w:`${50+u*17%35}%`,h:2})]},u))}),o.jsx(en,{w:s*.7,h:Math.min(32,r*.1),radius:6,style:{background:"var(--agd-bar)"}})]})}function d5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",padding:10,gap:8},children:[o.jsx("span",{style:{fontSize:18,lineHeight:1,color:"var(--agd-stroke)",fontFamily:"serif"},children:"“"}),o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:4},children:[o.jsx(se,{w:"90%",h:2}),o.jsx(se,{w:"75%",h:2}),o.jsx(se,{w:"60%",h:2})]}),o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:6},children:[o.jsx(bl,{size:20}),o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:2},children:[o.jsx(se,{w:60,h:3,strong:!0}),o.jsx(se,{w:40,h:2})]})]})]})}function f5({width:s,height:r}){return o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",height:"100%",gap:r*.08},children:[o.jsx(se,{w:s*.5,h:Math.max(4,r*.05),strong:!0}),o.jsx(se,{w:s*.35}),o.jsx(en,{w:Math.min(140,s*.25),h:Math.min(32,r*.15),radius:6,style:{marginTop:r*.04,background:"var(--agd-bar)"}})]})}function _5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",borderRadius:6,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",padding:"0 10px",gap:8},children:[o.jsx("div",{style:{width:16,height:16,borderRadius:"50%",border:"1.5px solid var(--agd-bar-strong)",display:"flex",alignItems:"center",justifyContent:"center",flexShrink:0},children:o.jsx("div",{style:{width:2,height:6,background:"var(--agd-bar-strong)",borderRadius:1}})}),o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:"40%",h:3,strong:!0}),o.jsx(se,{w:"70%",h:2})]})]})}function m5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",background:"var(--agd-fill)",display:"flex",alignItems:"center",justifyContent:"center",gap:8,padding:"0 12px"},children:[o.jsx(se,{w:s*.4,h:3,strong:!0}),o.jsx(en,{w:60,h:Math.min(24,r*.6),radius:4})]})}function h5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",gap:r*.06},children:[o.jsx(se,{w:s*.5,h:2}),o.jsx(se,{w:s*.4,h:Math.max(8,r*.18),strong:!0}),o.jsx(se,{w:s*.3,h:2})]})}function g5({width:s,height:r}){const c=Math.max(3,Math.min(5,Math.floor(s/100))),u=Math.min(12,r*.35);return o.jsx("div",{style:{display:"flex",alignItems:"center",justifyContent:"space-between",height:"100%",padding:"0 8px"},children:Array.from({length:c},(f,d)=>o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:0,flex:1},children:[o.jsx("div",{style:{width:u,height:u,borderRadius:"50%",border:"1.5px solid var(--agd-stroke)",background:d===0?"var(--agd-bar)":"transparent",flexShrink:0}}),d<c-1&&o.jsx("div",{style:{flex:1,height:1,background:"var(--agd-stroke)",margin:"0 4px"}})]},d))})}function p5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",borderRadius:4,border:"1px solid var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",justifyContent:"center",gap:4,padding:"0 6px"},children:[o.jsx(se,{w:Math.max(16,s*.5),h:2,strong:!0}),o.jsx("div",{style:{width:8,height:8,borderRadius:"50%",border:"1px solid var(--agd-stroke)",flexShrink:0}})]})}function y5({width:s,height:r}){const u=Math.min(r*.7,s/7.5);return o.jsx("div",{style:{display:"flex",alignItems:"center",justifyContent:"center",height:"100%",gap:u*.2},children:Array.from({length:5},(f,d)=>o.jsx("svg",{width:u,height:u,viewBox:"0 0 16 16",fill:"none",children:o.jsx("path",{d:"M8 1.5l2 4 4.5.7-3.25 3.1.75 4.5L8 11.4l-4 2.4.75-4.5L1.5 6.2 6 5.5z",stroke:"var(--agd-stroke)",strokeWidth:"0.8",fill:d<3?"var(--agd-bar)":"none"})},d))})}function x5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",position:"relative",borderRadius:4,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",overflow:"hidden"},children:[o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",style:{position:"absolute",inset:0},children:[o.jsx("line",{x1:0,y1:r*.3,x2:s,y2:r*.7,stroke:"var(--agd-stroke)",strokeWidth:"0.5",opacity:".2"}),o.jsx("line",{x1:0,y1:r*.6,x2:s,y2:r*.2,stroke:"var(--agd-stroke)",strokeWidth:"0.5",opacity:".15"}),o.jsx("line",{x1:s*.4,y1:0,x2:s*.6,y2:r,stroke:"var(--agd-stroke)",strokeWidth:"0.5",opacity:".15"})]}),o.jsx("div",{style:{position:"absolute",left:"50%",top:"40%",transform:"translate(-50%, -100%)"},children:o.jsxs("svg",{width:"16",height:"22",viewBox:"0 0 16 22",fill:"none",children:[o.jsx("path",{d:"M8 0C3.6 0 0 3.6 0 8c0 6 8 14 8 14s8-8 8-14c0-4.4-3.6-8-8-8z",fill:"var(--agd-bar)",opacity:".4"}),o.jsx("circle",{cx:"8",cy:"8",r:"3",fill:"var(--agd-fill)"})]})})]})}function b5({width:s,height:r}){const c=Math.max(3,Math.min(5,Math.floor(r/60)));return o.jsxs("div",{style:{display:"flex",height:"100%",padding:"8px 0"},children:[o.jsx("div",{style:{width:16,display:"flex",flexDirection:"column",alignItems:"center"},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",flex:1},children:[o.jsx(bl,{size:8}),f<c-1&&o.jsx("div",{style:{flex:1,width:1,background:"var(--agd-stroke)"}})]},f))}),o.jsx("div",{style:{flex:1,display:"flex",flexDirection:"column",justifyContent:"space-around",paddingLeft:8},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:`${35+f*13%25}%`,h:3,strong:!0}),o.jsx(se,{w:`${50+f*17%30}%`,h:2})]},f))})]})}function v5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",borderRadius:8,border:"2px dashed var(--agd-stroke)",display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",gap:r*.06},children:[o.jsxs("svg",{width:"24",height:"24",viewBox:"0 0 24 24",fill:"none",children:[o.jsx("path",{d:"M12 16V4m0 0l-4 4m4-4l4 4",stroke:"var(--agd-stroke)",strokeWidth:"1.5"}),o.jsx("path",{d:"M4 17v2a1 1 0 001 1h14a1 1 0 001-1v-2",stroke:"var(--agd-stroke)",strokeWidth:"1.5"})]}),o.jsx(se,{w:s*.4,h:2}),o.jsx(se,{w:s*.25,h:2})]})}function w5({width:s,height:r}){const c=Math.max(3,Math.min(8,Math.floor(r/20)));return o.jsxs("div",{style:{height:"100%",borderRadius:6,background:"var(--agd-fill)",border:"1px solid var(--agd-stroke)",padding:8,display:"flex",flexDirection:"column",gap:4},children:[o.jsxs("div",{style:{display:"flex",gap:3,marginBottom:4},children:[o.jsx(bl,{size:6}),o.jsx(bl,{size:6}),o.jsx(bl,{size:6})]}),Array.from({length:c},(u,f)=>o.jsx("div",{style:{display:"flex",gap:6,paddingLeft:f>0&&f<c-1?12:0},children:o.jsx(se,{w:`${25+f*23%50}%`,h:2,strong:f===0})},f))]})}function S5({width:s,height:r}){const f=Math.min((s-16)/7,(r-40)/6);return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsxs("div",{style:{display:"flex",alignItems:"center",justifyContent:"space-between",padding:"6px 8px"},children:[o.jsx("span",{style:{fontSize:8,color:"var(--agd-stroke)"},children:"‹"}),o.jsx(se,{w:s*.3,h:3,strong:!0}),o.jsx("span",{style:{fontSize:8,color:"var(--agd-stroke)"},children:"›"})]}),o.jsxs("div",{style:{display:"grid",gridTemplateColumns:"repeat(7, 1fr)",gap:2,padding:"0 4px",flex:1},children:[Array.from({length:7},(d,m)=>o.jsx("div",{style:{display:"flex",alignItems:"center",justifyContent:"center",height:f*.6},children:o.jsx(se,{w:f*.5,h:2})},`h${m}`)),Array.from({length:35},(d,m)=>o.jsx("div",{style:{display:"flex",alignItems:"center",justifyContent:"center",height:f},children:o.jsx("div",{style:{width:f*.6,height:f*.6,borderRadius:"50%",background:m===12?"var(--agd-bar)":"transparent",display:"flex",alignItems:"center",justifyContent:"center"},children:o.jsx("div",{style:{width:2,height:2,borderRadius:1,background:"var(--agd-bar-strong)",opacity:m===12?1:.3}})})},m))]})]})}function j5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",borderRadius:8,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",padding:"0 10px",gap:8},children:[o.jsx(bl,{size:Math.min(32,r*.55)}),o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:"50%",h:3,strong:!0}),o.jsx(se,{w:"75%",h:2})]}),o.jsx(se,{w:30,h:2})]})}function k5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column"},children:[o.jsx("div",{style:{height:"50%",background:"var(--agd-fill)",borderBottom:"1px dashed var(--agd-stroke)"}}),o.jsxs("div",{style:{flex:1,padding:10,display:"flex",flexDirection:"column",gap:5},children:[o.jsx(se,{w:"65%",h:4,strong:!0}),o.jsx(se,{w:"40%",h:3}),o.jsx("div",{style:{flex:1}}),o.jsxs("div",{style:{display:"flex",alignItems:"center",justifyContent:"space-between"},children:[o.jsx(se,{w:"30%",h:5,strong:!0}),o.jsx(en,{w:Math.min(70,s*.3),h:26,radius:4,style:{background:"var(--agd-bar)"}})]})]})]})}function C5({width:s,height:r}){const c=Math.min(48,r*.3);return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",gap:r*.06},children:[o.jsx(bl,{size:c}),o.jsx(se,{w:s*.45,h:4,strong:!0}),o.jsx(se,{w:s*.3,h:2}),o.jsxs("div",{style:{display:"flex",gap:s*.08,marginTop:r*.04},children:[o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",gap:2},children:[o.jsx(se,{w:20,h:3,strong:!0}),o.jsx(se,{w:28,h:2})]}),o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",gap:2},children:[o.jsx(se,{w:20,h:3,strong:!0}),o.jsx(se,{w:28,h:2})]}),o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",gap:2},children:[o.jsx(se,{w:20,h:3,strong:!0}),o.jsx(se,{w:28,h:2})]})]})]})}function N5({width:s,height:r}){const c=Math.max(s*.6,80),u=Math.max(3,Math.floor(r/40));return o.jsxs("div",{style:{height:"100%",display:"flex"},children:[o.jsx("div",{style:{width:s-c,background:"var(--agd-fill)",opacity:.3}}),o.jsxs("div",{style:{flex:1,borderLeft:"1px solid var(--agd-stroke)",display:"flex",flexDirection:"column",padding:s*.04},children:[o.jsxs("div",{style:{display:"flex",justifyContent:"space-between",alignItems:"center",marginBottom:r*.06},children:[o.jsx(se,{w:c*.4,h:4,strong:!0}),o.jsx("div",{style:{width:12,height:12,border:"1px solid var(--agd-stroke)",borderRadius:3}})]}),Array.from({length:u},(f,d)=>o.jsx("div",{style:{padding:"6px 0"},children:o.jsx(se,{w:`${50+d*17%35}%`,h:2,strong:d===0})},d))]})]})}function M5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center"},children:[o.jsxs("div",{style:{flex:1,width:"100%",borderRadius:8,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",padding:10,display:"flex",flexDirection:"column",gap:5},children:[o.jsx(se,{w:"70%",h:3,strong:!0}),o.jsx(se,{w:"90%",h:2}),o.jsx(se,{w:"60%",h:2})]}),o.jsx("div",{style:{width:10,height:10,background:"var(--agd-fill)",border:"1px dashed var(--agd-stroke)",borderTop:"none",borderLeft:"none",transform:"rotate(45deg)",marginTop:-6}})]})}function E5({width:s,height:r}){const c=Math.min(r*.7,s*.3);return o.jsxs("div",{style:{height:"100%",display:"flex",alignItems:"center",gap:s*.08},children:[o.jsx(en,{w:c,h:c,radius:c*.25}),o.jsx(se,{w:s*.45,h:Math.max(4,r*.2),strong:!0})]})}function T5({width:s,height:r}){const c=Math.max(2,Math.min(5,Math.floor(r/56)));return o.jsx("div",{style:{display:"flex",flexDirection:"column",height:"100%"},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{borderBottom:"1px solid var(--agd-stroke)",padding:"8px 6px",display:"flex",alignItems:"center",justifyContent:"space-between",flex:f===0?2:1},children:[o.jsxs("div",{style:{display:"flex",alignItems:"center",gap:6},children:[o.jsx("span",{style:{fontSize:9,fontWeight:700,color:"var(--agd-stroke)"},children:"Q"}),o.jsx(se,{w:s*(.3+f*13%25/100),h:3,strong:!0})]}),o.jsx("span",{style:{fontSize:8,color:"var(--agd-stroke)"},children:f===0?"▼":"▶"})]},f))})}function R5({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(s/120))),u=Math.max(1,Math.min(3,Math.floor(r/120)));return o.jsx("div",{style:{display:"grid",gridTemplateColumns:`repeat(${c}, 1fr)`,gridTemplateRows:`repeat(${u}, 1fr)`,gap:4,height:"100%"},children:Array.from({length:c*u},(f,d)=>o.jsx("div",{style:{borderRadius:4,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",position:"relative",overflow:"hidden"},children:o.jsxs("svg",{width:"100%",height:"100%",viewBox:"0 0 100 100",preserveAspectRatio:"none",fill:"none",children:[o.jsx("line",{x1:"0",y1:"0",x2:"100",y2:"100",stroke:"var(--agd-stroke)",strokeWidth:"0.5"}),o.jsx("line",{x1:"100",y1:"0",x2:"0",y2:"100",stroke:"var(--agd-stroke)",strokeWidth:"0.5"})]})},d))})}function z5({width:s,height:r}){const c=Math.min(s,r);return o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:[o.jsx("rect",{x:"1",y:(r-c+2)/2,width:c-2,height:c-2,rx:c*.15,stroke:"var(--agd-stroke)",strokeWidth:"1.5"}),o.jsx("path",{d:`M${c*.25} ${r/2}l${c*.2} ${c*.2} ${c*.3}-${c*.35}`,stroke:"var(--agd-bar)",strokeWidth:"1.5",fill:"none",strokeLinecap:"round",strokeLinejoin:"round"})]})}function A5({width:s,height:r}){const c=Math.min(s,r)/2-1;return o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:[o.jsx("circle",{cx:s/2,cy:r/2,r:c,stroke:"var(--agd-stroke)",strokeWidth:"1.5"}),o.jsx("circle",{cx:s/2,cy:r/2,r:c*.45,fill:"var(--agd-bar)"})]})}function O5({width:s,height:r}){const c=Math.max(2,r*.12),u=Math.min(r*.35,10),f=s*.55;return o.jsxs("div",{style:{height:"100%",display:"flex",alignItems:"center",position:"relative"},children:[o.jsx("div",{style:{width:"100%",height:c,borderRadius:c/2,background:"var(--agd-fill)",border:"1px solid var(--agd-stroke)",position:"relative"},children:o.jsx("div",{style:{width:f,height:"100%",borderRadius:c/2,background:"var(--agd-bar)"}})}),o.jsx("div",{style:{position:"absolute",left:f-u,width:u*2,height:u*2,borderRadius:"50%",border:"1.5px solid var(--agd-stroke)",background:"var(--agd-fill)"}})]})}function D5({width:s,height:r}){const c=Math.min(36,r*.15),u=7,f=4,d=Math.min((s-16)/u,(r-c-40)/(f+1));return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",gap:4},children:[o.jsxs("div",{style:{height:c,borderRadius:4,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",padding:"0 8px",justifyContent:"space-between"},children:[o.jsx(se,{w:"40%",h:2}),o.jsxs("svg",{width:"12",height:"12",viewBox:"0 0 16 16",fill:"none",children:[o.jsx("rect",{x:"2",y:"3",width:"12",height:"11",rx:"1",stroke:"var(--agd-stroke)",strokeWidth:"1"}),o.jsx("line",{x1:"2",y1:"6",x2:"14",y2:"6",stroke:"var(--agd-stroke)",strokeWidth:"0.5"})]})]}),o.jsxs("div",{style:{flex:1,borderRadius:6,border:"1px dashed var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",flexDirection:"column"},children:[o.jsxs("div",{style:{display:"flex",alignItems:"center",justifyContent:"space-between",padding:"4px 6px"},children:[o.jsx("span",{style:{fontSize:7,color:"var(--agd-stroke)"},children:"‹"}),o.jsx(se,{w:s*.25,h:2,strong:!0}),o.jsx("span",{style:{fontSize:7,color:"var(--agd-stroke)"},children:"›"})]}),o.jsx("div",{style:{display:"grid",gridTemplateColumns:`repeat(${u}, 1fr)`,gap:1,padding:"0 4px",flex:1},children:Array.from({length:u*f},(m,g)=>o.jsx("div",{style:{display:"flex",alignItems:"center",justifyContent:"center",height:d},children:o.jsx("div",{style:{width:d*.5,height:d*.5,borderRadius:"50%",background:g===10?"var(--agd-bar)":"transparent"},children:o.jsx("div",{style:{width:"100%",height:"100%",display:"flex",alignItems:"center",justifyContent:"center"},children:o.jsx("div",{style:{width:1.5,height:1.5,borderRadius:1,background:"var(--agd-bar-strong)",opacity:g===10?1:.25}})})})},g))})]})]})}function L5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",gap:r*.08,padding:4},children:[o.jsx("div",{style:{width:"100%",height:r*.2,borderRadius:4,background:"var(--agd-fill)"}}),o.jsx("div",{style:{width:"70%",height:Math.max(6,r*.1),borderRadius:3,background:"var(--agd-fill)"}}),o.jsx("div",{style:{width:"90%",height:Math.max(4,r*.06),borderRadius:3,background:"var(--agd-fill)"}}),o.jsx("div",{style:{width:"50%",height:Math.max(4,r*.06),borderRadius:3,background:"var(--agd-fill)"}})]})}function B5({width:s,height:r}){return o.jsx("div",{style:{height:"100%",display:"flex",alignItems:"center",gap:6},children:o.jsxs("div",{style:{height:"100%",flex:1,borderRadius:r/2,border:"1px solid var(--agd-stroke)",background:"var(--agd-fill)",display:"flex",alignItems:"center",padding:`0 ${r*.3}px`,gap:4},children:[o.jsx(se,{w:"60%",h:2,strong:!0}),o.jsx("div",{style:{width:Math.max(6,r*.3),height:Math.max(6,r*.3),borderRadius:"50%",border:"1px solid var(--agd-stroke)",flexShrink:0,marginLeft:"auto"}})]})})}function $5({width:s,height:r}){const c=Math.min(s,r);return o.jsx("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:o.jsx("path",{d:`M${s/2} ${(r-c)/2+c*.1}l${c*.12} ${c*.25} ${c*.28} ${c*.04}-${c*.2} ${c*.2} ${c*.05} ${c*.28}-${c*.25}-${c*.12}-${c*.25} ${c*.12} ${c*.05}-${c*.28}-${c*.2}-${c*.2} ${c*.28}-${c*.04}z`,stroke:"var(--agd-stroke)",strokeWidth:"1",fill:"var(--agd-fill)"})})}function U5({width:s,height:r}){const c=Math.min(s,r)/2-2;return o.jsxs("svg",{width:"100%",height:"100%",viewBox:`0 0 ${s} ${r}`,fill:"none",children:[o.jsx("circle",{cx:s/2,cy:r/2,r:c,stroke:"var(--agd-stroke)",strokeWidth:"1.5",opacity:".2"}),o.jsx("path",{d:`M${s/2} ${r/2-c}a${c} ${c} 0 0 1 ${c} ${c}`,stroke:"var(--agd-bar-strong)",strokeWidth:"1.5",strokeLinecap:"round"})]})}function H5({width:s,height:r}){const c=Math.min(36,r*.25,s*.12),u=Math.max(1,Math.min(3,Math.floor(r/80)));return o.jsx("div",{style:{display:"flex",flexDirection:"column",height:"100%",justifyContent:"space-around",padding:8},children:Array.from({length:u},(f,d)=>o.jsxs("div",{style:{display:"flex",gap:s*.04,alignItems:"flex-start"},children:[o.jsx(en,{w:c,h:c,radius:c*.25}),o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:4},children:[o.jsx(se,{w:`${40+d*13%20}%`,h:3,strong:!0}),o.jsx(se,{w:`${60+d*17%25}%`,h:2})]})]},d))})}function Y5({width:s,height:r}){const c=Math.max(2,Math.min(4,Math.floor(s/120))),u=Math.min(36,r*.25);return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center",gap:r*.06,padding:r*.06},children:[o.jsx(se,{w:s*.3,h:4,strong:!0}),o.jsx("div",{style:{display:"flex",gap:s*.06,justifyContent:"center",flex:1,alignItems:"center"},children:Array.from({length:c},(f,d)=>o.jsxs("div",{style:{display:"flex",flexDirection:"column",alignItems:"center",gap:6},children:[o.jsx(bl,{size:u}),o.jsx(se,{w:s*.12,h:3,strong:!0}),o.jsx(se,{w:s*.08,h:2})]},d))})]})}function q5({width:s,height:r}){const c=Math.max(2,Math.min(3,Math.floor(r/80)));return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",alignItems:"center",padding:s*.06,gap:r*.04},children:[o.jsx(se,{w:s*.5,h:Math.max(5,r*.04),strong:!0}),o.jsx(se,{w:s*.35,h:2}),o.jsx("div",{style:{width:"100%",display:"flex",flexDirection:"column",gap:r*.03,marginTop:r*.04},children:Array.from({length:c},(u,f)=>o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:Math.min(60,s*.2),h:2}),o.jsx(en,{w:"100%",h:Math.min(32,r*.1),radius:4})]},f))}),o.jsx(en,{w:"100%",h:Math.min(36,r*.12),radius:6,style:{marginTop:r*.03,background:"var(--agd-bar)"}}),o.jsx(se,{w:s*.4,h:2})]})}function I5({width:s,height:r}){return o.jsxs("div",{style:{height:"100%",display:"flex",flexDirection:"column",padding:s*.04,gap:r*.03},children:[o.jsx(se,{w:s*.4,h:4,strong:!0}),o.jsx(se,{w:s*.6,h:2}),o.jsxs("div",{style:{display:"flex",gap:6,marginTop:r*.03},children:[o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:50,h:2}),o.jsx(en,{w:"100%",h:Math.min(28,r*.1),radius:4})]}),o.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:40,h:2}),o.jsx(en,{w:"100%",h:Math.min(28,r*.1),radius:4})]})]}),o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:3},children:[o.jsx(se,{w:50,h:2}),o.jsx(en,{w:"100%",h:Math.min(28,r*.1),radius:4})]}),o.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:3,flex:1},children:[o.jsx(se,{w:60,h:2}),o.jsx(en,{w:"100%",h:"100%",radius:4})]}),o.jsx(en,{w:Math.min(120,s*.3),h:Math.min(30,r*.1),radius:6,style:{alignSelf:"flex-end",background:"var(--agd-bar)"}})]})}var X5={navigation:zv,hero:Av,sidebar:Ov,footer:Dv,modal:Lv,card:Bv,text:$v,image:Uv,table:Hv,list:Yv,button:qv,input:Iv,form:Xv,tabs:Qv,avatar:Gv,badge:Vv,header:Fv,section:Wv,grid:Kv,dropdown:Zv,toggle:Jv,search:Pv,toast:e5,progress:t5,chart:n5,video:l5,tooltip:o5,breadcrumb:s5,pagination:a5,divider:r5,accordion:i5,carousel:c5,pricing:u5,testimonial:d5,cta:f5,alert:_5,banner:m5,stat:h5,stepper:g5,tag:p5,rating:y5,map:x5,timeline:b5,fileUpload:v5,codeBlock:w5,calendar:S5,notification:j5,productCard:k5,profile:C5,drawer:N5,popover:M5,logo:E5,faq:T5,gallery:R5,checkbox:z5,radio:A5,slider:O5,datePicker:D5,skeleton:L5,chip:B5,icon:$5,spinner:U5,feature:H5,team:Y5,login:q5,contact:I5};function Q5({type:s,width:r,height:c,text:u}){const f=X5[s];return f?o.jsx("div",{style:{width:"100%",height:"100%",padding:8,position:"relative",pointerEvents:"none"},children:o.jsx(f,{width:r,height:c,text:u})}):o.jsx("div",{style:{width:"100%",height:"100%",display:"flex",alignItems:"center",justifyContent:"center"},children:o.jsx("span",{style:{fontSize:10,fontWeight:600,color:"var(--agd-text-3)",textTransform:"uppercase",letterSpacing:"0.06em",opacity:.5},children:s})})}var G5=`.styles-module__overlay___aWh-q svg[fill=none],
.styles-module__rearrangeOverlay___-3R3t svg[fill=none] {
  fill: none !important;
}
.styles-module__overlay___aWh-q svg[fill=none] :not([fill]),
.styles-module__rearrangeOverlay___-3R3t svg[fill=none] :not([fill]) {
  fill: none !important;
}

.styles-module__overlayExiting___iEmYr {
  opacity: 0 !important;
  transition: opacity 0.25s ease !important;
  pointer-events: none !important;
}

.styles-module__overlay___aWh-q {
  position: fixed;
  inset: 0;
  z-index: 99995;
  pointer-events: auto;
  cursor: default;
  animation: styles-module__overlayFadeIn___aECVy 0.15s ease;
  --agd-stroke: rgba(59, 130, 246, 0.35);
  --agd-fill: rgba(59, 130, 246, 0.06);
  --agd-bar: rgba(59, 130, 246, 0.18);
  --agd-bar-strong: rgba(59, 130, 246, 0.28);
  --agd-text-3: rgba(255, 255, 255, 0.6);
  --agd-surface: #fff;
}
.styles-module__overlay___aWh-q.styles-module__light___ORIft {
  --agd-surface: #fff;
}
.styles-module__overlay___aWh-q:not(.styles-module__light___ORIft) {
  --agd-surface: #141414;
}
.styles-module__overlay___aWh-q.styles-module__wireframe___itvQU {
  --agd-stroke: rgba(249, 115, 22, 0.35);
  --agd-fill: rgba(249, 115, 22, 0.06);
  --agd-bar: rgba(249, 115, 22, 0.18);
  --agd-bar-strong: rgba(249, 115, 22, 0.28);
}
.styles-module__overlay___aWh-q.styles-module__placing___45yD8 {
  cursor: crosshair;
}
.styles-module__overlay___aWh-q.styles-module__passthrough___xaFeE {
  pointer-events: none;
}

.styles-module__blankCanvas___t2Eue {
  position: fixed;
  inset: 0;
  z-index: 99994;
  background: #fff;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.25s ease;
}
.styles-module__blankCanvas___t2Eue.styles-module__visible___OKKqX {
  opacity: var(--canvas-opacity, 1);
  pointer-events: auto;
}
.styles-module__blankCanvas___t2Eue::after {
  content: "";
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(0, 0, 0, 0.08) 1px, transparent 1px);
  background-size: 24px 24px;
  background-position: 12px 12px;
  pointer-events: none;
  transition: opacity 0.2s ease;
}
.styles-module__blankCanvas___t2Eue.styles-module__gridActive___OZ-cf::after {
  opacity: 1;
  background-image: radial-gradient(circle, rgba(0, 0, 0, 0.22) 1px, transparent 1px);
}

.styles-module__paletteHeader___-Q5gQ {
  padding: 0 1rem 0.375rem;
}

.styles-module__paletteHeaderTitle___oHqZC {
  font-size: 0.8125rem;
  font-weight: 500;
  color: #fff;
  letter-spacing: -0.0094em;
}
.styles-module__light___ORIft .styles-module__paletteHeaderTitle___oHqZC {
  color: rgba(0, 0, 0, 0.85);
}

.styles-module__paletteHeaderDesc___6i74T {
  font-size: 0.6875rem;
  font-weight: 300;
  color: rgba(255, 255, 255, 0.45);
  margin-top: 2px;
  line-height: 14px;
}
.styles-module__light___ORIft .styles-module__paletteHeaderDesc___6i74T {
  color: rgba(0, 0, 0, 0.45);
}
.styles-module__paletteHeaderDesc___6i74T a {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: underline dotted;
  text-decoration-color: rgba(255, 255, 255, 0.2);
  text-underline-offset: 2px;
  transition: color 0.15s ease;
}
.styles-module__paletteHeaderDesc___6i74T a:hover {
  color: #fff;
}
.styles-module__light___ORIft .styles-module__paletteHeaderDesc___6i74T a {
  color: rgba(0, 0, 0, 0.6);
  text-decoration-color: rgba(0, 0, 0, 0.2);
}
.styles-module__light___ORIft .styles-module__paletteHeaderDesc___6i74T a:hover {
  color: rgba(0, 0, 0, 0.85);
}

.styles-module__wireframePurposeWrap___To-tS {
  display: grid;
  grid-template-rows: 1fr;
  transition: grid-template-rows 0.2s ease, opacity 0.15s ease;
  opacity: 1;
}
.styles-module__wireframePurposeWrap___To-tS.styles-module__collapsed___Ms9vS {
  grid-template-rows: 0fr;
  opacity: 0;
}

.styles-module__wireframePurposeInner___Lrahs {
  overflow: hidden;
}

.styles-module__wireframePurposeInput___7EtBN {
  display: block;
  width: calc(100% - 2rem);
  margin: 0.25rem 1rem 0.375rem;
  padding: 0.375rem 0.5rem;
  font-size: 0.8125rem;
  font-family: inherit;
  color: rgba(255, 255, 255, 0.85);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.375rem;
  resize: none;
  outline: none;
  transition: border-color 0.15s ease;
  letter-spacing: -0.0094em;
}
.styles-module__wireframePurposeInput___7EtBN::placeholder {
  color: rgba(255, 255, 255, 0.3);
}
.styles-module__wireframePurposeInput___7EtBN:focus {
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.05);
}
.styles-module__light___ORIft .styles-module__wireframePurposeInput___7EtBN {
  color: rgba(0, 0, 0, 0.7);
  background: rgba(0, 0, 0, 0.03);
  border-color: rgba(0, 0, 0, 0.1);
}
.styles-module__light___ORIft .styles-module__wireframePurposeInput___7EtBN::placeholder {
  color: rgba(0, 0, 0, 0.3);
}
.styles-module__light___ORIft .styles-module__wireframePurposeInput___7EtBN:focus {
  border-color: rgba(0, 0, 0, 0.25);
  background: rgba(0, 0, 0, 0.05);
}

.styles-module__canvasToggle___-QqSy {
  width: calc(100% - 2rem);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  margin: 0.25rem 1rem 0.25rem;
  padding: 0.375rem 0.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  border: 1px dashed rgba(255, 255, 255, 0.1);
  background: transparent;
  transition: background 0.15s ease, border-color 0.15s ease;
}
.styles-module__canvasToggle___-QqSy:hover {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.15);
}
.styles-module__canvasToggle___-QqSy.styles-module__active___hosp7 {
  background: #f97316;
  border-color: transparent;
  border-style: solid;
  box-shadow: none;
}
.styles-module__light___ORIft .styles-module__canvasToggle___-QqSy {
  border-color: rgba(0, 0, 0, 0.08);
}
.styles-module__light___ORIft .styles-module__canvasToggle___-QqSy:hover {
  background: rgba(0, 0, 0, 0.02);
  border-color: rgba(0, 0, 0, 0.12);
}
.styles-module__light___ORIft .styles-module__canvasToggle___-QqSy.styles-module__active___hosp7 {
  background: #f97316;
  border-color: transparent;
  border-style: solid;
  box-shadow: none;
}

.styles-module__canvasToggleIcon___7pJ82 {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.35);
}
.styles-module__active___hosp7 .styles-module__canvasToggleIcon___7pJ82 {
  color: rgba(255, 255, 255, 0.85);
}
.styles-module__light___ORIft .styles-module__canvasToggleIcon___7pJ82 {
  color: rgba(0, 0, 0, 0.25);
}
.styles-module__light___ORIft .styles-module__active___hosp7 .styles-module__canvasToggleIcon___7pJ82 {
  color: rgba(255, 255, 255, 0.85);
}

.styles-module__canvasToggleLabel___OanpY {
  font-size: 0.8125rem;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.6);
  letter-spacing: -0.0094em;
}
.styles-module__active___hosp7 .styles-module__canvasToggleLabel___OanpY {
  color: #fff;
}
.styles-module__light___ORIft .styles-module__canvasToggleLabel___OanpY {
  color: rgba(0, 0, 0, 0.5);
}
.styles-module__light___ORIft .styles-module__active___hosp7 .styles-module__canvasToggleLabel___OanpY {
  color: #fff;
}

.styles-module__placement___zcxv8 {
  position: absolute;
  border: 1.5px dashed rgba(59, 130, 246, 0.4);
  border-radius: 6px;
  background: rgba(59, 130, 246, 0.08);
  cursor: grab;
  transition: box-shadow 0.15s, border-color 0.15s, opacity 0.15s ease, transform 0.15s ease;
  -webkit-user-select: none;
  user-select: none;
  pointer-events: auto;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  animation: styles-module__placementEnter___TdRhf 0.25s cubic-bezier(0.34, 1.2, 0.64, 1);
}
.styles-module__placement___zcxv8:active {
  cursor: grabbing;
}
.styles-module__placement___zcxv8:hover {
  border-color: rgba(59, 130, 246, 0.5);
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.12);
}
.styles-module__placement___zcxv8.styles-module__selected___6yrp6 {
  border-color: #3c82f7;
  border-style: solid;
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15), 0 2px 8px rgba(59, 130, 246, 0.15);
}
.styles-module__placement___zcxv8.styles-module__selected___6yrp6:hover {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15), 0 2px 8px rgba(59, 130, 246, 0.15);
}
.styles-module__wireframe___itvQU .styles-module__placement___zcxv8 {
  border-color: rgba(249, 115, 22, 0.4);
  background: rgba(249, 115, 22, 0.08);
}
.styles-module__wireframe___itvQU .styles-module__placement___zcxv8:hover {
  border-color: rgba(249, 115, 22, 0.5);
  background: rgba(249, 115, 22, 0.1);
  box-shadow: 0 2px 8px rgba(249, 115, 22, 0.12);
}
.styles-module__wireframe___itvQU .styles-module__placement___zcxv8.styles-module__selected___6yrp6 {
  border-color: #f97316;
  background: rgba(249, 115, 22, 0.1);
  box-shadow: 0 0 0 2px rgba(249, 115, 22, 0.15), 0 2px 8px rgba(249, 115, 22, 0.15);
}
.styles-module__wireframe___itvQU .styles-module__placement___zcxv8.styles-module__selected___6yrp6:hover {
  box-shadow: 0 0 0 2px rgba(249, 115, 22, 0.15), 0 2px 8px rgba(249, 115, 22, 0.15);
}
.styles-module__placement___zcxv8.styles-module__dragging___le6KZ {
  opacity: 0.85;
  z-index: 50;
}
.styles-module__placement___zcxv8.styles-module__exiting___YrM8F {
  opacity: 0;
  transform: scale(0.97);
  pointer-events: none;
  animation: none;
  transition: opacity 0.2s ease, transform 0.2s cubic-bezier(0.32, 0.72, 0, 1);
}

.styles-module__placementContent___f64A4 {
  width: 100%;
  height: 100%;
  overflow: hidden;
  pointer-events: none;
}

.styles-module__placementLabel___0KvWl {
  position: absolute;
  top: -18px;
  left: 0;
  font-size: 10px;
  font-weight: 600;
  color: rgba(59, 130, 246, 0.7);
  white-space: nowrap;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  text-shadow: 0 0 4px rgba(255, 255, 255, 0.8), 0 0 8px rgba(255, 255, 255, 0.5);
}
.styles-module__selected___6yrp6 .styles-module__placementLabel___0KvWl {
  color: #3c82f7;
}
.styles-module__wireframe___itvQU .styles-module__placementLabel___0KvWl {
  color: rgba(249, 115, 22, 0.7);
}
.styles-module__wireframe___itvQU .styles-module__selected___6yrp6 .styles-module__placementLabel___0KvWl {
  color: #f97316;
}

.styles-module__placementAnnotation___78pTr {
  position: absolute;
  bottom: -18px;
  left: 0;
  right: 0;
  font-weight: 450;
  color: rgba(0, 0, 0, 0.5);
  font-size: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  text-shadow: 0 0 4px rgba(255, 255, 255, 0.9), 0 0 8px rgba(255, 255, 255, 0.6);
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.styles-module__placementAnnotation___78pTr.styles-module__annotationVisible___mrUyA {
  opacity: 1;
  transform: translateY(0);
}

.styles-module__sectionAnnotation___aUIs0 {
  position: absolute;
  bottom: -18px;
  left: 0;
  right: 0;
  font-weight: 450;
  color: rgba(59, 130, 246, 0.6);
  font-size: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  text-shadow: 0 0 4px rgba(255, 255, 255, 0.9), 0 0 8px rgba(255, 255, 255, 0.6);
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.styles-module__sectionAnnotation___aUIs0.styles-module__annotationVisible___mrUyA {
  opacity: 1;
  transform: translateY(0);
}

.styles-module__handle___Ikbxm {
  position: absolute;
  width: 8px;
  height: 8px;
  background: #fff;
  border: 1.5px solid #3c82f7;
  border-radius: 2px;
  z-index: 12;
  box-shadow: 0 0 0 0.5px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.12);
  opacity: 0;
  transform: scale(0.3);
  pointer-events: none;
  will-change: opacity, transform;
  transition: opacity 0.2s ease-out, transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.styles-module__placement___zcxv8:hover .styles-module__handle___Ikbxm, .styles-module__sectionOutline___s0hy-:hover .styles-module__handle___Ikbxm, .styles-module__ghostOutline___po-kO:hover .styles-module__handle___Ikbxm, .styles-module__placement___zcxv8:active .styles-module__handle___Ikbxm, .styles-module__sectionOutline___s0hy-:active .styles-module__handle___Ikbxm, .styles-module__ghostOutline___po-kO:active .styles-module__handle___Ikbxm, .styles-module__selected___6yrp6 .styles-module__handle___Ikbxm {
  opacity: 1;
  transform: scale(1);
  pointer-events: auto;
}
.styles-module__sectionOutline___s0hy- .styles-module__handle___Ikbxm {
  border-color: inherit;
}
.styles-module__wireframe___itvQU .styles-module__handle___Ikbxm {
  border-color: #f97316;
}

.styles-module__handleNw___4TMIj {
  top: -4px;
  left: -4px;
  cursor: nw-resize;
}

.styles-module__handleNe___mnsTh {
  top: -4px;
  right: -4px;
  cursor: ne-resize;
}

.styles-module__handleSe___oSFnk {
  bottom: -4px;
  right: -4px;
  cursor: se-resize;
}

.styles-module__handleSw___pi--Z {
  bottom: -4px;
  left: -4px;
  cursor: sw-resize;
}

.styles-module__handleN___aBA-Q,
.styles-module__handleE___0hM5u,
.styles-module__handleS___JjDRv,
.styles-module__handleW___ERWGQ {
  opacity: 0 !important;
  pointer-events: none !important;
}

.styles-module__edgeHandle___XxXdT {
  position: absolute;
  z-index: 11;
  display: flex;
  align-items: center;
  justify-content: center;
}
.styles-module__edgeHandle___XxXdT::after {
  content: "";
  position: absolute;
  border-radius: 4px;
  background: #3c82f7;
}
.styles-module__wireframe___itvQU .styles-module__edgeHandle___XxXdT::after {
  background: #f97316;
}
.styles-module__edgeHandle___XxXdT::after {
  opacity: 0;
  transition: opacity 0.1s ease, transform 0.1s ease;
  transform: scale(0.8);
}
.styles-module__edgeHandle___XxXdT:hover::after {
  opacity: 0.85;
  transform: scale(1);
}
.styles-module__edgeHandle___XxXdT svg {
  position: relative;
  z-index: 1;
  opacity: 0;
  transition: opacity 0.1s ease;
  filter: drop-shadow(0 0 2px var(--agd-surface));
}
.styles-module__edgeHandle___XxXdT:hover svg {
  opacity: 1;
}

.styles-module__edgeN___-JJDj,
.styles-module__edgeS___66lMX {
  left: 12px;
  right: 12px;
  height: 12px;
  cursor: n-resize;
}
.styles-module__edgeN___-JJDj::after,
.styles-module__edgeS___66lMX::after {
  width: 24px;
  height: 4px;
}

.styles-module__edgeN___-JJDj {
  top: -6px;
}

.styles-module__edgeS___66lMX {
  bottom: -6px;
  cursor: s-resize;
}

.styles-module__edgeE___1bGDa,
.styles-module__edgeW___lHQNo {
  top: 12px;
  bottom: 12px;
  width: 12px;
  cursor: e-resize;
}
.styles-module__edgeE___1bGDa::after,
.styles-module__edgeW___lHQNo::after {
  width: 4px;
  height: 24px;
}

.styles-module__edgeE___1bGDa {
  right: -6px;
}

.styles-module__edgeW___lHQNo {
  left: -6px;
  cursor: w-resize;
}

.styles-module__deleteButton___LkGCb {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  color: rgba(0, 0, 0, 0.35);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  line-height: 1;
  z-index: 15;
  pointer-events: none;
  opacity: 0;
  transform: scale(0.8);
  will-change: opacity, transform;
  transition: opacity 0.2s ease-out, transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), background 0.12s ease, color 0.12s ease, border-color 0.12s ease, box-shadow 0.12s ease;
}
.styles-module__placement___zcxv8:hover .styles-module__deleteButton___LkGCb, .styles-module__selected___6yrp6 .styles-module__deleteButton___LkGCb, .styles-module__sectionOutline___s0hy-:hover .styles-module__deleteButton___LkGCb, .styles-module__sectionOutline___s0hy-.styles-module__selected___6yrp6 .styles-module__deleteButton___LkGCb, .styles-module__ghostOutline___po-kO:hover .styles-module__deleteButton___LkGCb, .styles-module__ghostOutline___po-kO.styles-module__selected___6yrp6 .styles-module__deleteButton___LkGCb {
  opacity: 1;
  transform: scale(1);
  pointer-events: auto;
}
.styles-module__deleteButton___LkGCb:hover {
  background: #ef4444;
  color: #fff;
  border-color: #ef4444;
  box-shadow: 0 1px 4px rgba(239, 68, 68, 0.3);
  transform: scale(1.1);
}
.styles-module__overlay___aWh-q:not(.styles-module__light___ORIft) .styles-module__deleteButton___LkGCb, .styles-module__rearrangeOverlay___-3R3t:not(.styles-module__light___ORIft) .styles-module__deleteButton___LkGCb {
  background: rgba(40, 40, 40, 0.9);
  border-color: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.5);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
}
.styles-module__overlay___aWh-q:not(.styles-module__light___ORIft) .styles-module__deleteButton___LkGCb:hover, .styles-module__rearrangeOverlay___-3R3t:not(.styles-module__light___ORIft) .styles-module__deleteButton___LkGCb:hover {
  background: #ef4444;
  color: #fff;
  border-color: #ef4444;
}

.styles-module__drawBox___BrVAa {
  position: fixed;
  pointer-events: none;
  z-index: 99996;
  border: 2px solid #3c82f7;
  border-radius: 6px;
  background: rgba(59, 130, 246, 0.15);
}

.styles-module__selectBox___Iu8kB {
  position: fixed;
  pointer-events: none;
  z-index: 99996;
  border: 1px dashed #3c82f7;
  background: rgba(59, 130, 246, 0.08);
  border-radius: 2px;
}

.styles-module__sizeIndicator___7zJ4y {
  position: fixed;
  pointer-events: none;
  z-index: 100001;
  font-size: 10px;
  color: #fff;
  background: #3c82f7;
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
  font-weight: 500;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

.styles-module__guideLine___DUQY2 {
  pointer-events: none;
  z-index: 100001;
  background: #f0f;
  opacity: 0.5;
}

.styles-module__dragPreview___onPbU {
  position: fixed;
  z-index: 100002;
  pointer-events: none;
  border: 1.5px dashed #3c82f7;
  border-radius: 6px;
  background: rgba(59, 130, 246, 0.1);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 600;
  color: #3c82f7;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.15);
  transition: width 0.08s ease, height 0.08s ease, opacity 0.08s ease;
}

.styles-module__dragPreviewWireframe___jsg0G {
  border-color: #f97316;
  background: rgba(249, 115, 22, 0.1);
  color: #f97316;
  box-shadow: 0 4px 16px rgba(249, 115, 22, 0.15);
}

.styles-module__palette___C7iSH {
  position: absolute;
  right: 5px;
  bottom: calc(100% + 0.5rem);
  width: 256px;
  overflow: hidden;
  background: #1c1c1c;
  border: none;
  border-radius: 1rem;
  padding: 13px 0 16px;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(0, 0, 0, 0.04);
  z-index: 100001;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  cursor: default;
  opacity: 0;
  filter: blur(5px);
}
.styles-module__palette___C7iSH .styles-module__paletteItem___6TlnA,
.styles-module__palette___C7iSH .styles-module__paletteItemLabel___6ncO4,
.styles-module__palette___C7iSH .styles-module__paletteSectionTitle___PqnjX,
.styles-module__palette___C7iSH .styles-module__paletteFooter___QYnAG {
  transition: background 0.25s ease, color 0.25s ease, border-color 0.25s ease;
}
.styles-module__palette___C7iSH {
  opacity: 0;
  transform: translateY(var(--panel-offset-y, 4px)) scale(0.98);
  transform-origin: var(--panel-origin, bottom right);
  filter: blur(2px);
  pointer-events: none;
  visibility: hidden;
  transition: opacity 120ms cubic-bezier(0.25, 0.46, 0.45, 0.94), transform 120ms cubic-bezier(0.25, 0.46, 0.45, 0.94), filter 120ms cubic-bezier(0.25, 0.46, 0.45, 0.94);
}
.styles-module__palette___C7iSH[data-panel-present=true] {
  visibility: visible;
}
.styles-module__palette___C7iSH[data-panel-open=true] {
  opacity: 1;
  transform: translateY(0) scale(1);
  filter: blur(0);
  pointer-events: auto;
  transition-duration: 160ms;
}
@media (prefers-reduced-motion: reduce) {
  .styles-module__palette___C7iSH {
    transition: none;
    transform: none;
    filter: none;
  }
}
.styles-module__palette___C7iSH.styles-module__light___ORIft {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08), 0 4px 16px rgba(0, 0, 0, 0.06), 0 0 0 1px rgba(0, 0, 0, 0.04);
}

.styles-module__paletteSection___V8DEA {
  padding: 0 1rem;
}
.styles-module__paletteSection___V8DEA + .styles-module__paletteSection___V8DEA {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
}
.styles-module__light___ORIft .styles-module__paletteSection___V8DEA + .styles-module__paletteSection___V8DEA {
  border-top-color: rgba(0, 0, 0, 0.07);
}

.styles-module__paletteSectionTitle___PqnjX {
  font-size: 0.6875rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.5);
  letter-spacing: -0.0094em;
  padding: 0 0 3px 3px;
}
.styles-module__light___ORIft .styles-module__paletteSectionTitle___PqnjX {
  color: rgba(0, 0, 0, 0.4);
}

.styles-module__paletteItem___6TlnA {
  width: 100%;
  text-align: left;
  background: transparent;
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.25rem;
  margin-bottom: 1px;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.15s ease, border-color 0.15s ease;
  border: 1px solid transparent;
  -webkit-user-select: none;
  user-select: none;
  min-height: 24px;
}
.styles-module__paletteItem___6TlnA:hover {
  background: rgba(255, 255, 255, 0.1);
}
.styles-module__paletteItem___6TlnA.styles-module__active___hosp7 {
  background: #3c82f7;
  border-color: transparent;
}
.styles-module__paletteItem___6TlnA.styles-module__wireframe___itvQU.styles-module__active___hosp7 {
  background: #f97316;
}
.styles-module__light___ORIft .styles-module__paletteItem___6TlnA:hover {
  background: rgba(0, 0, 0, 0.05);
}
.styles-module__light___ORIft .styles-module__paletteItem___6TlnA.styles-module__active___hosp7 {
  background: #3c82f7;
  border-color: transparent;
}
.styles-module__light___ORIft .styles-module__paletteItem___6TlnA.styles-module__wireframe___itvQU.styles-module__active___hosp7 {
  background: #f97316;
}

.styles-module__paletteItemIcon___0NPQK {
  width: 20px;
  height: 16px;
  border-radius: 2px;
  border: 1px dashed rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
  color: rgba(255, 255, 255, 0.45);
}
.styles-module__paletteItemIcon___0NPQK svg {
  display: block;
  width: 20px;
  height: 16px;
}
.styles-module__active___hosp7 .styles-module__paletteItemIcon___0NPQK {
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}
.styles-module__light___ORIft .styles-module__paletteItemIcon___0NPQK {
  border-color: rgba(0, 0, 0, 0.12);
  background: rgba(0, 0, 0, 0.02);
  color: rgba(0, 0, 0, 0.4);
}
.styles-module__light___ORIft .styles-module__active___hosp7 .styles-module__paletteItemIcon___0NPQK {
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

.styles-module__paletteItemLabel___6ncO4 {
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  letter-spacing: -0.0094em;
  line-height: 1;
  min-width: 0;
}
.styles-module__active___hosp7 .styles-module__paletteItemLabel___6ncO4 {
  color: #fff;
  font-weight: 600;
}
.styles-module__light___ORIft .styles-module__paletteItemLabel___6ncO4 {
  color: rgba(0, 0, 0, 0.7);
}
.styles-module__light___ORIft .styles-module__active___hosp7 .styles-module__paletteItemLabel___6ncO4 {
  color: #fff;
  font-weight: 600;
}

.styles-module__placeScroll___7sClM {
  max-height: 240px;
  overflow-y: auto;
  overflow-x: hidden;
  padding-top: 0.25rem;
}
.styles-module__placeScroll___7sClM.styles-module__fadeTop___KT9tF {
  -webkit-mask-image: linear-gradient(to bottom, transparent 0, black 32px);
  mask-image: linear-gradient(to bottom, transparent 0, black 32px);
}
.styles-module__placeScroll___7sClM.styles-module__fadeBottom___x3ShT {
  -webkit-mask-image: linear-gradient(to bottom, black calc(100% - 32px), transparent 100%);
  mask-image: linear-gradient(to bottom, black calc(100% - 32px), transparent 100%);
}
.styles-module__placeScroll___7sClM.styles-module__fadeTop___KT9tF.styles-module__fadeBottom___x3ShT {
  -webkit-mask-image: linear-gradient(to bottom, transparent 0, black 32px, black calc(100% - 32px), transparent 100%);
  mask-image: linear-gradient(to bottom, transparent 0, black 32px, black calc(100% - 32px), transparent 100%);
}
.styles-module__placeScroll___7sClM::-webkit-scrollbar {
  width: 3px;
}
.styles-module__placeScroll___7sClM::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 2px;
}
.styles-module__light___ORIft .styles-module__placeScroll___7sClM::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
}

.styles-module__paletteFooterWrap___71-fI {
  display: grid;
  grid-template-rows: 1fr;
  transition: grid-template-rows 0.25s cubic-bezier(0.32, 0.72, 0, 1);
}
.styles-module__paletteFooterWrap___71-fI.styles-module__footerHidden___fJUik {
  grid-template-rows: 0fr;
}

.styles-module__paletteFooterInnerContent___VC26h {
  opacity: 1;
  transform: translateY(0);
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.styles-module__footerHidden___fJUik .styles-module__paletteFooterInnerContent___VC26h {
  opacity: 0;
  transform: translateY(4px);
}

.styles-module__paletteFooterInner___dfylY {
  overflow: hidden;
}

.styles-module__paletteFooter___QYnAG {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 24px;
  padding: 0 1rem;
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
}
.styles-module__light___ORIft .styles-module__paletteFooter___QYnAG {
  border-top-color: rgba(0, 0, 0, 0.07);
}

.styles-module__paletteFooterCount___D3Fia {
  font-size: 0.8125rem;
  font-weight: 400;
  letter-spacing: -0.0094em;
  color: rgba(255, 255, 255, 0.5);
}
.styles-module__light___ORIft .styles-module__paletteFooterCount___D3Fia {
  color: rgba(0, 0, 0, 0.5);
}

.styles-module__paletteFooterClear___ybBoa {
  font-size: 0.8125rem;
  font-weight: 400;
  letter-spacing: -0.0094em;
  color: rgba(255, 255, 255, 0.5);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  font-family: inherit;
  transition: color 0.15s ease;
}
.styles-module__paletteFooterClear___ybBoa:hover {
  color: rgba(255, 255, 255, 0.7);
}
.styles-module__light___ORIft .styles-module__paletteFooterClear___ybBoa {
  color: rgba(0, 0, 0, 0.5);
}
.styles-module__light___ORIft .styles-module__paletteFooterClear___ybBoa:hover {
  color: rgba(0, 0, 0, 0.6);
}

.styles-module__paletteFooterActions___fLzv8 {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.styles-module__rollingWrap___S75jM {
  display: inline-block;
  overflow: hidden;
  height: 1.15em;
  position: relative;
  vertical-align: bottom;
}

.styles-module__rollingNum___1RKDx {
  position: absolute;
  left: 0;
  top: 0;
}

.styles-module__exitUp___AFDRW {
  animation: styles-module__numExitUp___FRQqx 0.25s cubic-bezier(0.32, 0.72, 0, 1) forwards;
}

.styles-module__enterUp___CPlXb {
  animation: styles-module__numEnterUp___2Yd-w 0.25s cubic-bezier(0.32, 0.72, 0, 1) forwards;
}

.styles-module__exitDown___-1yAy {
  animation: styles-module__numExitDown___xm5by 0.25s cubic-bezier(0.32, 0.72, 0, 1) forwards;
}

.styles-module__enterDown___DDuFR {
  animation: styles-module__numEnterDown___hpxBk 0.25s cubic-bezier(0.32, 0.72, 0, 1) forwards;
}

@keyframes styles-module__numExitUp___FRQqx {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-110%);
    opacity: 0;
  }
}
@keyframes styles-module__numEnterUp___2Yd-w {
  from {
    transform: translateY(110%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
@keyframes styles-module__numExitDown___xm5by {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(110%);
    opacity: 0;
  }
}
@keyframes styles-module__numEnterDown___hpxBk {
  from {
    transform: translateY(-110%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
.styles-module__rearrangeOverlay___-3R3t {
  position: fixed;
  inset: 0;
  z-index: 99995;
  pointer-events: none;
  cursor: default;
  -webkit-user-select: none;
  user-select: none;
  animation: styles-module__overlayFadeIn___aECVy 0.15s ease;
}

.styles-module__hoverHighlight___8eT-v {
  position: fixed;
  pointer-events: none;
  z-index: 99994;
  border: 2px dashed rgba(59, 130, 246, 0.5);
  border-radius: 4px;
  background: rgba(59, 130, 246, 0.06);
  animation: styles-module__highlightFadeIn___Lg7KY 0.12s ease;
}

.styles-module__sectionOutline___s0hy- {
  position: fixed;
  border: 2px solid;
  border-radius: 4px;
  cursor: grab;
}
.styles-module__sectionOutline___s0hy-:active {
  cursor: grabbing;
}
.styles-module__sectionOutline___s0hy- {
  transition: box-shadow 0.15s, border-color 0.3s, background-color 0.3s, border-style 0s;
  -webkit-user-select: none;
  user-select: none;
  pointer-events: auto;
  animation: styles-module__sectionEnter___-8BXT 0.2s ease;
}
.styles-module__sectionOutline___s0hy-:hover {
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.1), 0 4px 12px rgba(0, 0, 0, 0.15);
}
.styles-module__sectionOutline___s0hy-.styles-module__selected___6yrp6 {
  border-style: solid;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15), 0 2px 8px rgba(59, 130, 246, 0.15);
}
.styles-module__sectionOutline___s0hy-.styles-module__selected___6yrp6:hover {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15), 0 2px 8px rgba(59, 130, 246, 0.15);
}
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6) {
  border: 1.5px dashed rgba(150, 150, 150, 0.35);
  background-color: transparent !important;
  box-shadow: none;
}
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6):hover {
  border-color: rgba(150, 150, 150, 0.6);
  box-shadow: none;
}
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6) .styles-module__sectionLabel___F80HQ {
  opacity: 0;
  transition: opacity 0.15s ease;
}
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6):hover .styles-module__sectionLabel___F80HQ {
  opacity: 1;
}
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6) .styles-module__movedBadge___s8z-q,
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6) .styles-module__sectionDimensions___RcJSL {
  opacity: 0;
  transition: opacity 0.15s ease;
}
.styles-module__sectionOutline___s0hy-.styles-module__settled___b5U5o:not(.styles-module__selected___6yrp6):hover .styles-module__sectionDimensions___RcJSL {
  opacity: 1;
}
.styles-module__sectionOutline___s0hy-.styles-module__exiting___YrM8F {
  opacity: 0;
  transform: scale(0.97);
  pointer-events: none;
  animation: none;
  transition: opacity 0.2s ease, transform 0.2s cubic-bezier(0.32, 0.72, 0, 1);
}

.styles-module__sectionLabel___F80HQ {
  position: absolute;
  top: 4px;
  left: 4px;
  font-size: 10px;
  font-weight: 600;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  white-space: nowrap;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  max-width: calc(100% - 8px);
  overflow: hidden;
  text-overflow: ellipsis;
}

.styles-module__movedBadge___s8z-q {
  position: absolute;
  bottom: 22px;
  right: 4px;
  font-size: 9px;
  font-weight: 700;
  color: #fff;
  background: #22c55e;
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  opacity: 0;
  transform: scale(0.8);
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.styles-module__movedBadge___s8z-q.styles-module__badgeVisible___npbdS {
  opacity: 1;
  transform: scale(1);
  transition: opacity 0.2s cubic-bezier(0.34, 1.2, 0.64, 1), transform 0.2s cubic-bezier(0.34, 1.2, 0.64, 1);
}

.styles-module__resizedBadge___u51V8 {
  background: #3c82f7;
  bottom: 40px;
}

.styles-module__sectionDimensions___RcJSL {
  position: absolute;
  bottom: 4px;
  right: 4px;
  font-size: 9px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.7);
  background: rgba(0, 0, 0, 0.5);
  padding: 1px 5px;
  border-radius: 3px;
  white-space: nowrap;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
}
.styles-module__light___ORIft .styles-module__sectionDimensions___RcJSL {
  color: rgba(0, 0, 0, 0.5);
  background: rgba(255, 255, 255, 0.7);
}

.styles-module__wireframeNotice___4GJyB {
  position: fixed;
  bottom: 16px;
  left: 24px;
  z-index: 99995;
  font-size: 9.5px;
  font-weight: 400;
  color: rgba(0, 0, 0, 0.4);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  pointer-events: auto;
  animation: styles-module__overlayFadeIn___aECVy 0.3s ease;
  line-height: 1.5;
  max-width: 280px;
}

.styles-module__wireframeOpacityRow___CJXzi {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.styles-module__wireframeOpacityLabel___afkfT {
  font-size: 9px;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.32);
  letter-spacing: 0.02em;
  white-space: nowrap;
  -webkit-user-select: none;
  user-select: none;
}

.styles-module__wireframeOpacitySlider___YcoEs {
  -webkit-appearance: none;
  appearance: none;
  width: 56px;
  height: 4px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s ease;
}
.styles-module__wireframeOpacitySlider___YcoEs:hover {
  background: rgba(0, 0, 0, 0.13);
}
.styles-module__wireframeOpacitySlider___YcoEs::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #f97316;
  cursor: pointer;
  transition: background 0.15s ease;
}
.styles-module__wireframeOpacitySlider___YcoEs::-webkit-slider-thumb:hover {
  background: rgb(88.0082041185%, 37.39404381%, 2.2663056855%);
}
.styles-module__wireframeOpacitySlider___YcoEs::-moz-range-thumb {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #f97316;
  border: none;
  cursor: pointer;
}
.styles-module__wireframeOpacitySlider___YcoEs::-moz-range-track {
  background: rgba(0, 0, 0, 0.08);
  height: 4px;
  border-radius: 2px;
}

.styles-module__wireframeNoticeTitleRow___PJqyG {
  display: flex;
  align-items: center;
  gap: 0;
  margin-bottom: 2px;
}

.styles-module__wireframeNoticeTitle___okr08 {
  font-weight: 600;
  color: rgba(0, 0, 0, 0.55);
}

.styles-module__wireframeNoticeDivider___PNKQ6 {
  width: 1px;
  height: 8px;
  background: rgba(0, 0, 0, 0.12);
  margin: 0 8px;
  flex-shrink: 0;
}

.styles-module__wireframeStartOver___YFk-I {
  font-size: 9.5px;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.35);
  cursor: pointer;
  background: none;
  border: none;
  padding: 0;
  font-family: inherit;
  text-decoration: none;
  transition: color 0.12s ease;
  white-space: nowrap;
}
.styles-module__wireframeStartOver___YFk-I:hover {
  color: rgba(0, 0, 0, 0.6);
}

.styles-module__ghostOutline___po-kO {
  position: fixed;
  border: 1.5px dashed rgba(59, 130, 246, 0.4);
  border-radius: 4px;
  background: rgba(59, 130, 246, 0.04);
  cursor: grab;
  opacity: 0.5;
  -webkit-user-select: none;
  user-select: none;
  pointer-events: auto;
  animation: styles-module__ghostEnter___EC3Mb 0.25s ease;
  transition: box-shadow 0.15s, border-color 0.3s, opacity 0.25s;
}
.styles-module__ghostOutline___po-kO:active {
  cursor: grabbing;
}
.styles-module__ghostOutline___po-kO:hover {
  opacity: 0.7;
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.1), 0 4px 12px rgba(0, 0, 0, 0.08);
}
.styles-module__ghostOutline___po-kO.styles-module__selected___6yrp6 {
  opacity: 1;
  border-style: solid;
  border-width: 2px;
  border-color: #3c82f7;
  background: rgba(59, 130, 246, 0.08);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15), 0 2px 8px rgba(59, 130, 246, 0.15);
}
.styles-module__ghostOutline___po-kO.styles-module__exiting___YrM8F {
  opacity: 0;
  transform: scale(0.97);
  pointer-events: none;
  animation: none;
  transition: opacity 0.2s ease, transform 0.2s cubic-bezier(0.32, 0.72, 0, 1);
}

.styles-module__ghostBadge___tsQUK {
  position: absolute;
  bottom: calc(100% + 4px);
  left: -1px;
  font-size: 9px;
  font-weight: 600;
  color: rgba(59, 130, 246, 0.9);
  background: rgba(59, 130, 246, 0.08);
  border: 1px solid rgba(59, 130, 246, 0.2);
  padding: 1px 5px;
  border-radius: 3px;
  white-space: nowrap;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  letter-spacing: 0.02em;
  line-height: 1.2;
  animation: styles-module__badgeSlideIn___typJ7 0.2s ease both;
}

@keyframes styles-module__badgeSlideIn___typJ7 {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
.styles-module__ghostBadgeExtra___6CVoD {
  display: inline;
  animation: styles-module__badgeExtraIn___i4W8F 0.2s ease both;
}

@keyframes styles-module__badgeExtraIn___i4W8F {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
.styles-module__originalOutline___Y6DD1 {
  position: fixed;
  border: 1.5px dashed rgba(150, 150, 150, 0.3);
  border-radius: 4px;
  background: transparent;
  pointer-events: none;
  -webkit-user-select: none;
  user-select: none;
  animation: styles-module__sectionEnter___-8BXT 0.2s ease;
}

.styles-module__originalLabel___HqI9g {
  position: absolute;
  top: 4px;
  left: 4px;
  font-size: 9px;
  font-weight: 500;
  color: rgba(150, 150, 150, 0.5);
  padding: 1px 6px;
  border-radius: 3px;
  white-space: nowrap;
  pointer-events: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background: rgba(150, 150, 150, 0.08);
}

.styles-module__connectorSvg___Lovld {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  pointer-events: none;
  z-index: 99996;
}

.styles-module__connectorLine___XeWh- {
  transition: opacity 0.2s ease;
  animation: styles-module__connectorDraw___8sK5I 0.3s ease both;
}

.styles-module__connectorDot___yvf7C {
  transform-box: fill-box;
  transform-origin: center;
  animation: styles-module__connectorDotIn___NwTUq 0.25s cubic-bezier(0.34, 1.56, 0.64, 1) 0.15s both;
}

@keyframes styles-module__connectorDraw___8sK5I {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
@keyframes styles-module__connectorDotIn___NwTUq {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
.styles-module__connectorExiting___2lLOs {
  animation: styles-module__connectorOut___5QoPl 0.2s ease forwards;
}
.styles-module__connectorExiting___2lLOs .styles-module__connectorDot___yvf7C {
  animation: styles-module__connectorDotOut___FEq7e 0.2s ease forwards;
}

@keyframes styles-module__connectorOut___5QoPl {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}
@keyframes styles-module__connectorDotOut___FEq7e {
  from {
    transform: scale(1);
    opacity: 1;
  }
  to {
    transform: scale(0);
    opacity: 0;
  }
}
@keyframes styles-module__placementEnter___TdRhf {
  from {
    opacity: 0;
    transform: scale(0.85);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
@keyframes styles-module__sectionEnter___-8BXT {
  from {
    opacity: 0;
    transform: scale(0.96);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
@keyframes styles-module__highlightFadeIn___Lg7KY {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
@keyframes styles-module__overlayFadeIn___aECVy {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
@keyframes styles-module__ghostEnter___EC3Mb {
  from {
    opacity: 0;
    transform: scale(0.96);
  }
  to {
    opacity: 0.6;
    transform: scale(1);
  }
}
.styles-module__canvasToggle___-QqSy:focus-visible,
.styles-module__paletteItem___6TlnA:focus-visible {
  outline: 2px solid var(--agentation-color-accent);
  outline-offset: -2px;
}`,ee={overlay:"styles-module__overlay___aWh-q",rearrangeOverlay:"styles-module__rearrangeOverlay___-3R3t",overlayExiting:"styles-module__overlayExiting___iEmYr",overlayFadeIn:"styles-module__overlayFadeIn___aECVy",light:"styles-module__light___ORIft",wireframe:"styles-module__wireframe___itvQU",placing:"styles-module__placing___45yD8",passthrough:"styles-module__passthrough___xaFeE",blankCanvas:"styles-module__blankCanvas___t2Eue",visible:"styles-module__visible___OKKqX",gridActive:"styles-module__gridActive___OZ-cf",paletteHeader:"styles-module__paletteHeader___-Q5gQ",paletteHeaderTitle:"styles-module__paletteHeaderTitle___oHqZC",paletteHeaderDesc:"styles-module__paletteHeaderDesc___6i74T",wireframePurposeWrap:"styles-module__wireframePurposeWrap___To-tS",collapsed:"styles-module__collapsed___Ms9vS",wireframePurposeInner:"styles-module__wireframePurposeInner___Lrahs",wireframePurposeInput:"styles-module__wireframePurposeInput___7EtBN",canvasToggle:"styles-module__canvasToggle___-QqSy",active:"styles-module__active___hosp7",canvasToggleIcon:"styles-module__canvasToggleIcon___7pJ82",canvasToggleLabel:"styles-module__canvasToggleLabel___OanpY",placement:"styles-module__placement___zcxv8",placementEnter:"styles-module__placementEnter___TdRhf",selected:"styles-module__selected___6yrp6",dragging:"styles-module__dragging___le6KZ",exiting:"styles-module__exiting___YrM8F",placementContent:"styles-module__placementContent___f64A4",placementLabel:"styles-module__placementLabel___0KvWl",placementAnnotation:"styles-module__placementAnnotation___78pTr",annotationVisible:"styles-module__annotationVisible___mrUyA",sectionAnnotation:"styles-module__sectionAnnotation___aUIs0",handle:"styles-module__handle___Ikbxm",sectionOutline:"styles-module__sectionOutline___s0hy-",ghostOutline:"styles-module__ghostOutline___po-kO",handleNw:"styles-module__handleNw___4TMIj",handleNe:"styles-module__handleNe___mnsTh",handleSe:"styles-module__handleSe___oSFnk",handleSw:"styles-module__handleSw___pi--Z",handleN:"styles-module__handleN___aBA-Q",handleE:"styles-module__handleE___0hM5u",handleS:"styles-module__handleS___JjDRv",handleW:"styles-module__handleW___ERWGQ",edgeHandle:"styles-module__edgeHandle___XxXdT",edgeN:"styles-module__edgeN___-JJDj",edgeS:"styles-module__edgeS___66lMX",edgeE:"styles-module__edgeE___1bGDa",edgeW:"styles-module__edgeW___lHQNo",deleteButton:"styles-module__deleteButton___LkGCb",drawBox:"styles-module__drawBox___BrVAa",selectBox:"styles-module__selectBox___Iu8kB",sizeIndicator:"styles-module__sizeIndicator___7zJ4y",guideLine:"styles-module__guideLine___DUQY2",dragPreview:"styles-module__dragPreview___onPbU",dragPreviewWireframe:"styles-module__dragPreviewWireframe___jsg0G",palette:"styles-module__palette___C7iSH",paletteItem:"styles-module__paletteItem___6TlnA",paletteItemLabel:"styles-module__paletteItemLabel___6ncO4",paletteSectionTitle:"styles-module__paletteSectionTitle___PqnjX",paletteFooter:"styles-module__paletteFooter___QYnAG",paletteSection:"styles-module__paletteSection___V8DEA",paletteItemIcon:"styles-module__paletteItemIcon___0NPQK",placeScroll:"styles-module__placeScroll___7sClM",fadeTop:"styles-module__fadeTop___KT9tF",fadeBottom:"styles-module__fadeBottom___x3ShT",paletteFooterWrap:"styles-module__paletteFooterWrap___71-fI",footerHidden:"styles-module__footerHidden___fJUik",paletteFooterInnerContent:"styles-module__paletteFooterInnerContent___VC26h",paletteFooterInner:"styles-module__paletteFooterInner___dfylY",paletteFooterCount:"styles-module__paletteFooterCount___D3Fia",paletteFooterClear:"styles-module__paletteFooterClear___ybBoa",paletteFooterActions:"styles-module__paletteFooterActions___fLzv8",rollingWrap:"styles-module__rollingWrap___S75jM",rollingNum:"styles-module__rollingNum___1RKDx",exitUp:"styles-module__exitUp___AFDRW",numExitUp:"styles-module__numExitUp___FRQqx",enterUp:"styles-module__enterUp___CPlXb",numEnterUp:"styles-module__numEnterUp___2Yd-w",exitDown:"styles-module__exitDown___-1yAy",numExitDown:"styles-module__numExitDown___xm5by",enterDown:"styles-module__enterDown___DDuFR",numEnterDown:"styles-module__numEnterDown___hpxBk",hoverHighlight:"styles-module__hoverHighlight___8eT-v",highlightFadeIn:"styles-module__highlightFadeIn___Lg7KY",sectionEnter:"styles-module__sectionEnter___-8BXT",settled:"styles-module__settled___b5U5o",sectionLabel:"styles-module__sectionLabel___F80HQ",movedBadge:"styles-module__movedBadge___s8z-q",sectionDimensions:"styles-module__sectionDimensions___RcJSL",badgeVisible:"styles-module__badgeVisible___npbdS",resizedBadge:"styles-module__resizedBadge___u51V8",wireframeNotice:"styles-module__wireframeNotice___4GJyB",wireframeOpacityRow:"styles-module__wireframeOpacityRow___CJXzi",wireframeOpacityLabel:"styles-module__wireframeOpacityLabel___afkfT",wireframeOpacitySlider:"styles-module__wireframeOpacitySlider___YcoEs",wireframeNoticeTitleRow:"styles-module__wireframeNoticeTitleRow___PJqyG",wireframeNoticeTitle:"styles-module__wireframeNoticeTitle___okr08",wireframeNoticeDivider:"styles-module__wireframeNoticeDivider___PNKQ6",wireframeStartOver:"styles-module__wireframeStartOver___YFk-I",ghostEnter:"styles-module__ghostEnter___EC3Mb",ghostBadge:"styles-module__ghostBadge___tsQUK",badgeSlideIn:"styles-module__badgeSlideIn___typJ7",ghostBadgeExtra:"styles-module__ghostBadgeExtra___6CVoD",badgeExtraIn:"styles-module__badgeExtraIn___i4W8F",originalOutline:"styles-module__originalOutline___Y6DD1",originalLabel:"styles-module__originalLabel___HqI9g",connectorSvg:"styles-module__connectorSvg___Lovld",connectorLine:"styles-module__connectorLine___XeWh-",connectorDraw:"styles-module__connectorDraw___8sK5I",connectorDot:"styles-module__connectorDot___yvf7C",connectorDotIn:"styles-module__connectorDotIn___NwTUq",connectorExiting:"styles-module__connectorExiting___2lLOs",connectorOut:"styles-module__connectorOut___5QoPl",connectorDotOut:"styles-module__connectorDotOut___FEq7e"},Ur=24,qu=5;function Y0(s,r,c,u,f){let d=1/0,m=1/0;const g=s.x,p=s.x+s.width,j=s.x+s.width/2,k=s.y,b=s.y+s.height,v=s.y+s.height/2,z=!u,E=z?[g,p,j]:[...u.left?[g]:[],...u.right?[p]:[]],$=z?[k,b,v]:[...u.top?[k]:[],...u.bottom?[b]:[]],B=[];for(const pe of r)c.has(pe.id)||B.push(pe);f&&B.push(...f);for(const pe of B){const V=pe.x,Ie=pe.x+pe.width,he=pe.x+pe.width/2,ue=pe.y,G=pe.y+pe.height,we=pe.y+pe.height/2;for(const fe of E)for(const Oe of[V,Ie,he]){const Se=Oe-fe;Math.abs(Se)<qu&&Math.abs(Se)<Math.abs(d)&&(d=Se)}for(const fe of $)for(const Oe of[ue,G,we]){const Se=Oe-fe;Math.abs(Se)<qu&&Math.abs(Se)<Math.abs(m)&&(m=Se)}}const F=Math.abs(d)<qu?d:0,Me=Math.abs(m)<qu?m:0,te=[],C=new Set,ie=g+F,de=p+F,W=j+F,Ee=k+Me,ne=b+Me,Ce=v+Me;for(const pe of B){const V=pe.x,Ie=pe.x+pe.width,he=pe.x+pe.width/2,ue=pe.y,G=pe.y+pe.height,we=pe.y+pe.height/2;for(const fe of[V,he,Ie])for(const Oe of[ie,W,de])if(Math.abs(Oe-fe)<.5){const Se=`x:${Math.round(fe)}`;C.has(Se)||(C.add(Se),te.push({axis:"x",pos:fe}))}for(const fe of[ue,we,G])for(const Oe of[Ee,Ce,ne])if(Math.abs(Oe-fe)<.5){const Se=`y:${Math.round(fe)}`;C.has(Se)||(C.add(Se),te.push({axis:"y",pos:fe}))}}return{dx:F,dy:Me,guides:te}}function q0(){return`dp-${Date.now()}-${Math.random().toString(36).slice(2,7)}`}function V5({placements:s,onChange:r,activeComponent:c,onActiveComponentChange:u,isDarkMode:f,exiting:d,onInteractionChange:m,className:g,passthrough:p,extraSnapRects:j,onSelectionChange:k,deselectSignal:b,onDragMove:v,onDragEnd:z,clearingPlacements:E,wireframe:$}){const[B,F]=h.useState(new Set),[Me,te]=h.useState(null),[C,ie]=h.useState(null),[de,W]=h.useState(null),[Ee,ne]=h.useState([]),[Ce,pe]=h.useState(null),[V,Ie]=h.useState(!1),he=h.useRef(!1),[ue,G]=h.useState(new Set),we=h.useRef(new Map),fe=h.useRef(null),Oe=h.useRef(null),Se=h.useRef(s);Se.current=s;const Qe=h.useRef(k);Qe.current=k;const Ut=h.useRef(v);Ut.current=v;const Ye=h.useRef(z);Ye.current=z;const Ne=h.useRef(b);h.useEffect(()=>{b!==Ne.current&&(Ne.current=b,F(new Set))},[b]),h.useEffect(()=>{E!=null&&E.length&&(F(re=>new Set([...re].filter(ye=>!E.some(Ge=>Ge.id===ye)))),Oe.current=null)},[E]),h.useEffect(()=>{const re=ye=>{const Ge=ye.composedPath()[0]||ye.target;if(!(Ge.tagName==="INPUT"||Ge.tagName==="TEXTAREA"||Ge.isContentEditable)){if((ye.key==="Backspace"||ye.key==="Delete")&&B.size>0){ye.preventDefault();const lt=new Set(B);G(lt),F(new Set),Ht(()=>{r(Se.current.filter(rt=>!lt.has(rt.id))),G(new Set)},180);return}if(["ArrowUp","ArrowDown","ArrowLeft","ArrowRight"].includes(ye.key)&&B.size>0){ye.preventDefault();const lt=ye.shiftKey?20:1,rt=ye.key==="ArrowLeft"?-lt:ye.key==="ArrowRight"?lt:0,it=ye.key==="ArrowUp"?-lt:ye.key==="ArrowDown"?lt:0;r(s.map(Fe=>B.has(Fe.id)?{...Fe,x:Math.max(0,Fe.x+rt),y:Math.max(0,Fe.y+it)}:Fe));return}if(ye.key==="Escape"){c?u(null):B.size>0&&F(new Set);return}}};return document.addEventListener("keydown",re),()=>document.removeEventListener("keydown",re)},[B,c,s,r,u]);const Xt=h.useCallback(re=>{if(re.button!==0||p||re.target.closest(`.${ee.placement}`))return;re.preventDefault(),re.stopPropagation();const Ge=window.scrollY,nt=re.clientX,lt=re.clientY;if(c){Oe.current="place",m==null||m(!0);let rt=!1,it=nt,Fe=lt;const Nt=oe=>{it=oe.clientX,Fe=oe.clientY;const O=Math.abs(it-nt),I=Math.abs(Fe-lt);if((O>5||I>5)&&(rt=!0),rt){const P=Math.min(nt,it),ae=Math.min(lt,Fe),Be=Math.abs(it-nt),ke=Math.abs(Fe-lt);te({x:P,y:ae,w:Be,h:ke}),W({x:oe.clientX+12,y:oe.clientY+12,text:`${Math.round(Be)} × ${Math.round(ke)}`})}},an=oe=>{window.removeEventListener("mousemove",Nt),window.removeEventListener("mouseup",an),te(null),W(null),Oe.current=null,m==null||m(!1);const O=Ae[c];let I,P,ae,Be;rt?(I=Math.min(nt,it),P=Math.min(lt,Fe)+Ge,ae=Math.max(Ur,Math.abs(it-nt)),Be=Math.max(Ur,Math.abs(Fe-lt))):(ae=O.width,Be=O.height,I=nt-ae/2,P=lt+Ge-Be/2),I=Math.max(0,I),P=Math.max(0,P);const ke={id:q0(),type:c,x:I,y:P,width:ae,height:Be,scrollY:Ge,timestamp:Date.now()},_e=[...s,ke];r(_e),F(new Set([ke.id])),u(null)};window.addEventListener("mousemove",Nt),window.addEventListener("mouseup",an)}else{re.shiftKey||F(new Set),Oe.current="select";let rt=!1;const it=Nt=>{const an=Math.abs(Nt.clientX-nt),oe=Math.abs(Nt.clientY-lt);if((an>4||oe>4)&&(rt=!0),rt){const O=Math.min(nt,Nt.clientX),I=Math.min(lt,Nt.clientY);ie({x:O,y:I,w:Math.abs(Nt.clientX-nt),h:Math.abs(Nt.clientY-lt)})}},Fe=Nt=>{if(window.removeEventListener("mousemove",it),window.removeEventListener("mouseup",Fe),Oe.current=null,rt){const an=Math.min(nt,Nt.clientX),oe=Math.min(lt,Nt.clientY)+Ge,O=Math.abs(Nt.clientX-nt),I=Math.abs(Nt.clientY-lt),P=new Set(re.shiftKey?B:new Set);for(const ae of s)ae.y-Ge,ae.x+ae.width>an&&ae.x<an+O&&ae.y+ae.height>oe&&ae.y<oe+I&&P.add(ae.id);F(P)}ie(null)};window.addEventListener("mousemove",it),window.addEventListener("mouseup",Fe)}},[c,p,s,r,B]),xe=h.useCallback((re,ye)=>{var ke;if(re.button!==0)return;const Ge=re.target;if(Ge.closest(`.${ee.handle}`)||Ge.closest(`.${ee.deleteButton}`))return;re.preventDefault(),re.stopPropagation();let nt;re.shiftKey?(nt=new Set(B),nt.has(ye)?nt.delete(ye):nt.add(ye)):B.has(ye)?nt=new Set(B):nt=new Set([ye]),F(nt),(nt.size!==B.size||[...nt].some(_e=>!B.has(_e)))&&((ke=Qe.current)==null||ke.call(Qe,nt,re.shiftKey));const rt=re.clientX,it=re.clientY,Fe=new Map;for(const _e of s)nt.has(_e.id)&&Fe.set(_e.id,{x:_e.x,y:_e.y});Oe.current="move",m==null||m(!0);let Nt=!1,an=!1,oe=s,O=0,I=0;const P=new Map;for(const _e of s)Fe.has(_e.id)&&P.set(_e.id,{w:_e.width,h:_e.height});const ae=_e=>{var Nn;const at=_e.clientX-rt,ot=_e.clientY-it;if((Math.abs(at)>2||Math.abs(ot)>2)&&(Nt=!0),!Nt)return;if(_e.altKey&&!an){an=!0;const vt=[];for(const Je of s)Fe.has(Je.id)&&vt.push({...Je,id:q0(),timestamp:Date.now()});oe=[...s,...vt]}let ct=1/0,Ke=1/0,Ze=-1/0,bt=-1/0;for(const[vt,Je]of Fe){const je=P.get(vt);je&&(ct=Math.min(ct,Je.x+at),Ke=Math.min(Ke,Je.y+ot),Ze=Math.max(Ze,Je.x+at+je.w),bt=Math.max(bt,Je.y+ot+je.h))}const Bt={x:ct,y:Ke,width:Ze-ct,height:bt-Ke},{dx:pn,dy:pt,guides:At}=Y0(Bt,oe,new Set(Fe.keys()),void 0,j);ne(At);const kt=at+pn,ht=ot+pt;O=kt,I=ht,r(oe.map(vt=>{const Je=Fe.get(vt.id);return Je?{...vt,x:Math.max(0,Je.x+kt),y:Math.max(0,Je.y+ht)}:vt})),(Nn=Ut.current)==null||Nn.call(Ut,kt,ht)},Be=()=>{var _e;window.removeEventListener("mousemove",ae),window.removeEventListener("mouseup",Be),Oe.current=null,m==null||m(!1),ne([]),(_e=Ye.current)==null||_e.call(Ye,O,I,Nt)};window.addEventListener("mousemove",ae),window.addEventListener("mouseup",Be)},[B,s,r,m]),We=h.useCallback((re,ye,Ge)=>{re.preventDefault(),re.stopPropagation();const nt=s.find(P=>P.id===ye);if(!nt)return;F(new Set([ye])),Oe.current="resize",m==null||m(!0);const lt=re.clientX,rt=re.clientY,it=nt.width,Fe=nt.height,Nt=nt.x,an=nt.y,oe={left:Ge.includes("w"),right:Ge.includes("e"),top:Ge.includes("n"),bottom:Ge.includes("s")},O=P=>{const ae=P.clientX-lt,Be=P.clientY-rt;let ke=it,_e=Fe,at=Nt,ot=an;Ge.includes("e")&&(ke=Math.max(Ur,it+ae)),Ge.includes("w")&&(ke=Math.max(Ur,it-ae),at=Nt+it-ke),Ge.includes("s")&&(_e=Math.max(Ur,Fe+Be)),Ge.includes("n")&&(_e=Math.max(Ur,Fe-Be),ot=an+Fe-_e);const ct={x:at,y:ot,width:ke,height:_e},{dx:Ke,dy:Ze,guides:bt}=Y0(ct,Se.current,new Set([ye]),oe,j);ne(bt),Ke!==0&&(oe.right?ke+=Ke:oe.left&&(at+=Ke,ke-=Ke)),Ze!==0&&(oe.bottom?_e+=Ze:oe.top&&(ot+=Ze,_e-=Ze)),r(Se.current.map(Bt=>Bt.id===ye?{...Bt,x:at,y:ot,width:ke,height:_e}:Bt)),W({x:P.clientX+12,y:P.clientY+12,text:`${Math.round(ke)} × ${Math.round(_e)}`})},I=()=>{window.removeEventListener("mousemove",O),window.removeEventListener("mouseup",I),W(null),Oe.current=null,m==null||m(!1),ne([])};window.addEventListener("mousemove",O),window.addEventListener("mouseup",I)},[s,r,m]),gn=h.useCallback(re=>{Oe.current=null,G(ye=>{const Ge=new Set(ye);return Ge.add(re),Ge}),F(ye=>{const Ge=new Set(ye);return Ge.delete(re),Ge}),Ht(()=>{r(Se.current.filter(ye=>ye.id!==re)),G(ye=>{const Ge=new Set(ye);return Ge.delete(re),Ge})},180)},[r]),Qt={hero:"Headline text",button:"Button label",badge:"Badge label",cta:"Call to action text",toast:"Notification message",modal:"Dialog title",card:"Card title",navigation:"Brand / nav items",tabs:"Tab labels",input:"Placeholder text",search:"Search placeholder",pricing:"Plan name or price",testimonial:"Quote text",alert:"Alert message",banner:"Banner text",tag:"Tag label",notification:"Notification message",stat:"Metric value",productCard:"Product name"},mn=h.useCallback(re=>{const ye=s.find(Ge=>Ge.id===re);ye&&(he.current=!!ye.text,pe(re),Ie(!1))},[s]),Re=h.useCallback(()=>{Ce&&(Ie(!0),Ht(()=>{pe(null),Ie(!1)},150))},[Ce]);h.useEffect(()=>{d&&Ce&&Re()},[d]);const Ue=h.useCallback(re=>{Ce&&(r(s.map(ye=>ye.id===Ce?{...ye,text:re.trim()||void 0}:ye)),Re())},[Ce,s,r,Re]),un=typeof window<"u"?window.scrollY:0,ul=["nw","ne","se","sw"],Cn=$?"#f97316":"#3c82f7",bn=[{dir:"n",cls:ee.edgeN,arrow:o.jsx("svg",{width:"8",height:"6",viewBox:"0 0 8 6",fill:"none",children:o.jsx("path",{d:"M4 0.5L1 4.5h6z",fill:Cn})})},{dir:"e",cls:ee.edgeE,arrow:o.jsx("svg",{width:"6",height:"8",viewBox:"0 0 6 8",fill:"none",children:o.jsx("path",{d:"M5.5 4L1.5 1v6z",fill:Cn})})},{dir:"s",cls:ee.edgeS,arrow:o.jsx("svg",{width:"8",height:"6",viewBox:"0 0 8 6",fill:"none",children:o.jsx("path",{d:"M4 5.5L1 1.5h6z",fill:Cn})})},{dir:"w",cls:ee.edgeW,arrow:o.jsx("svg",{width:"6",height:"8",viewBox:"0 0 6 8",fill:"none",children:o.jsx("path",{d:"M0.5 4L4.5 1v6z",fill:Cn})})}];return o.jsxs(o.Fragment,{children:[o.jsx("div",{ref:fe,className:`${ee.overlay} ${f?"":ee.light} ${c?ee.placing:""} ${p?ee.passthrough:""} ${d?ee.overlayExiting:""} ${$?ee.wireframe:""}${g?` ${g}`:""}`,"data-feedback-toolbar":!0,onMouseDown:Xt,children:s.map(re=>{var lt;const ye=B.has(re.id),Ge=((lt=mo[re.type])==null?void 0:lt.label)||re.type,nt=re.y-un;return o.jsxs("div",{"data-design-placement":re.id,className:`${ee.placement} ${ye?ee.selected:""} ${ue.has(re.id)||E!=null&&E.includes(re)?ee.exiting:""}`,style:{left:re.x,top:nt,width:re.width,height:re.height,position:"fixed"},onMouseDown:rt=>xe(rt,re.id),onDoubleClick:()=>mn(re.id),children:[o.jsx("span",{className:ee.placementLabel,children:Ge}),o.jsx("span",{className:`${ee.placementAnnotation} ${re.text?ee.annotationVisible:""}`,children:(re.text&&we.current.set(re.id,re.text),re.text||we.current.get(re.id)||"")}),o.jsx("div",{className:ee.placementContent,children:o.jsx(Q5,{type:re.type,width:re.width,height:re.height,text:re.text})}),o.jsx("div",{className:ee.deleteButton,onMouseDown:rt=>rt.stopPropagation(),onClick:()=>gn(re.id),children:"✕"}),ul.map(rt=>o.jsx("div",{className:`${ee.handle} ${ee[`handle${rt.charAt(0).toUpperCase()}${rt.slice(1)}`]}`,onMouseDown:it=>We(it,re.id,rt)},rt)),bn.map(({dir:rt,cls:it,arrow:Fe})=>o.jsx("div",{className:`${ee.edgeHandle} ${it}`,onMouseDown:Nt=>We(Nt,re.id,rt),children:Fe},rt))]},re.id)})}),Ce&&(()=>{var an;const re=s.find(oe=>oe.id===Ce);if(!re)return null;const ye=re.y-un,Ge=re.x+re.width/2,nt=ye-8,lt=ye+re.height+8,rt=nt>200,it=lt<window.innerHeight-100,Fe=Math.max(160,Math.min(window.innerWidth-160,Ge));let Nt;return rt?Nt={left:Fe,bottom:window.innerHeight-nt}:it?Nt={left:Fe,top:lt}:Nt={left:Fe,top:Math.max(80,window.innerHeight/2-80)},o.jsx(K_,{element:((an=mo[re.type])==null?void 0:an.label)||re.type,placeholder:Qt[re.type]||"Label or content text",initialValue:re.text??"",submitLabel:he.current?"Save":"Set",onSubmit:Ue,onCancel:Re,onDelete:he.current?()=>{Ue("")}:void 0,isExiting:V,lightMode:!f,style:Nt})})(),Me&&o.jsx("div",{className:ee.drawBox,style:{left:Me.x,top:Me.y,width:Me.w,height:Me.h},"data-feedback-toolbar":!0}),C&&o.jsx("div",{className:ee.selectBox,style:{left:C.x,top:C.y,width:C.w,height:C.h},"data-feedback-toolbar":!0}),de&&o.jsx("div",{className:ee.sizeIndicator,style:{left:de.x,top:de.y},"data-feedback-toolbar":!0,children:de.text}),Ee.map((re,ye)=>o.jsx("div",{className:ee.guideLine,style:re.axis==="x"?{position:"fixed",left:re.pos,top:0,width:1,bottom:0}:{position:"fixed",left:0,top:re.pos-un,right:0,height:1},"data-feedback-toolbar":!0},`${re.axis}-${re.pos}-${ye}`))]})}function ty(s,{keepMounted:r=!1,onExited:c}={}){const[u,f]=h.useState(r||s),d=h.useRef(null),m=h.useRef(c);return s&&!u&&f(!0),h.useLayoutEffect(()=>{m.current=c},[c]),h.useLayoutEffect(()=>{var k;const g=d.current;if(!g||g.dataset.panelOpen==="true"===s)return;getComputedStyle(g).opacity,g.dataset.panelPresent="true",g.dataset.panelOpen=String(s);let p=!1;const j=((k=g.getAnimations)==null?void 0:k.call(g))??[];return Promise.allSettled(j.map(b=>b.finished)).then(()=>{var b;p||s||(delete g.dataset.panelPresent,r||f(!1),(b=m.current)==null||b.call(m))}),()=>{p=!0}},[s,u,r]),{ref:d,mounted:u}}function F5(s){if(!s)return"";const r=s.scrollTop>2,c=s.scrollTop+s.clientHeight<s.scrollHeight-2;return`${r?ee.fadeTop:""} ${c?ee.fadeBottom:""}`}var w="currentColor",K="0.5";function W5({type:s}){switch(s){case"navigation":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"4",width:"18",height:"8",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2.5",y:"7",width:"3",height:"1.5",rx:".5",fill:w,opacity:".4"}),o.jsx("rect",{x:"7",y:"7",width:"2.5",height:"1.5",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"11",y:"7",width:"2.5",height:"1.5",rx:".5",fill:w,opacity:".25"})]});case"header":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"2",width:"18",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3",y:"5.5",width:"8",height:"2",rx:".5",fill:w,opacity:".35"}),o.jsx("rect",{x:"3",y:"9",width:"12",height:"1",rx:".5",fill:w,opacity:".15"})]});case"hero":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"1",width:"18",height:"14",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5",y:"5",width:"10",height:"1.5",rx:".5",fill:w,opacity:".35"}),o.jsx("rect",{x:"7",y:"8",width:"6",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"7.5",y:"10.5",width:"5",height:"2.5",rx:"1",stroke:w,strokeWidth:K})]});case"section":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"1",width:"18",height:"14",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3",y:"4",width:"6",height:"1",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"3",y:"6.5",width:"14",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"3",y:"9",width:"10",height:"1",rx:".5",fill:w,opacity:".15"})]});case"sidebar":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"1",width:"7",height:"14",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2.5",y:"4",width:"4",height:"1",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"2.5",y:"6.5",width:"3.5",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"2.5",y:"9",width:"4",height:"1",rx:".5",fill:w,opacity:".15"})]});case"footer":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"7",width:"18",height:"8",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3",y:"9.5",width:"4",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"9",y:"9.5",width:"4",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"15",y:"9.5",width:"3",height:"1",rx:".5",fill:w,opacity:".2"})]});case"modal":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"2",width:"14",height:"12",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5",y:"4.5",width:"7",height:"1",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"5",y:"7",width:"10",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"11",y:"11",width:"5",height:"2",rx:".75",stroke:w,strokeWidth:K})]});case"divider":return o.jsx("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:o.jsx("line",{x1:"2",y1:"8",x2:"18",y2:"8",stroke:w,strokeWidth:"0.5",opacity:".3"})});case"card":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"1",width:"16",height:"14",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2",y:"1",width:"16",height:"5.5",rx:"1",fill:w,opacity:".04"}),o.jsx("rect",{x:"4",y:"8.5",width:"8",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"4",y:"11",width:"11",height:"1",rx:".5",fill:w,opacity:".12"})]});case"text":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"4",width:"14",height:"1.5",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"2",y:"7",width:"11",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"2",y:"9.5",width:"13",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"2",y:"12",width:"8",height:"1",rx:".5",fill:w,opacity:".12"})]});case"image":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"16",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("line",{x1:"2",y1:"2",x2:"18",y2:"14",stroke:w,strokeWidth:".3",opacity:".25"}),o.jsx("line",{x1:"18",y1:"2",x2:"2",y2:"14",stroke:w,strokeWidth:".3",opacity:".25"})]});case"video":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"16",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("path",{d:"M8.5 5.5v5l4.5-2.5z",stroke:w,strokeWidth:K,fill:w,opacity:".15"})]});case"table":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"2",width:"18",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("line",{x1:"1",y1:"5.5",x2:"19",y2:"5.5",stroke:w,strokeWidth:".3",opacity:".25"}),o.jsx("line",{x1:"1",y1:"9",x2:"19",y2:"9",stroke:w,strokeWidth:".3",opacity:".25"}),o.jsx("line",{x1:"7",y1:"2",x2:"7",y2:"14",stroke:w,strokeWidth:".3",opacity:".25"}),o.jsx("line",{x1:"13",y1:"2",x2:"13",y2:"14",stroke:w,strokeWidth:".3",opacity:".25"})]});case"grid":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1.5",y:"2",width:"7",height:"5.5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"11.5",y:"2",width:"7",height:"5.5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"1.5",y:"9.5",width:"7",height:"5.5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"11.5",y:"9.5",width:"7",height:"5.5",rx:"1",stroke:w,strokeWidth:K})]});case"list":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"3.5",cy:"4.5",r:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6.5",y:"4",width:"10",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("circle",{cx:"3.5",cy:"8",r:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6.5",y:"7.5",width:"8",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("circle",{cx:"3.5",cy:"11.5",r:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6.5",y:"11",width:"11",height:"1",rx:".5",fill:w,opacity:".2"})]});case"chart":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"9",width:"2.5",height:"4",rx:".5",fill:w,opacity:".2"}),o.jsx("rect",{x:"7",y:"6",width:"2.5",height:"7",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"11",y:"3",width:"2.5",height:"10",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"15",y:"5",width:"2.5",height:"8",rx:".5",fill:w,opacity:".2"})]});case"accordion":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1.5",y:"2",width:"17",height:"4",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3",y:"3.5",width:"6",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"1.5",y:"7.5",width:"17",height:"3",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"1.5",y:"12",width:"17",height:"3",rx:"1",stroke:w,strokeWidth:K})]});case"carousel":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"2",width:"14",height:"10",rx:"1",stroke:w,strokeWidth:K}),o.jsx("path",{d:"M1.5 7L3 8.5 1.5 10",stroke:w,strokeWidth:K,opacity:".35"}),o.jsx("path",{d:"M18.5 7L17 8.5 18.5 10",stroke:w,strokeWidth:K,opacity:".35"}),o.jsx("circle",{cx:"8.5",cy:"14",r:".6",fill:w,opacity:".35"}),o.jsx("circle",{cx:"10",cy:"14",r:".6",fill:w,opacity:".15"}),o.jsx("circle",{cx:"11.5",cy:"14",r:".6",fill:w,opacity:".15"})]});case"button":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"5",width:"14",height:"6",rx:"2",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6.5",y:"7.5",width:"7",height:"1",rx:".5",fill:w,opacity:".25"})]});case"input":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"4",width:"5.5",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"2",y:"6.5",width:"16",height:"5.5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3.5",y:"8.5",width:"7",height:"1",rx:".5",fill:w,opacity:".12"})]});case"search":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"4.5",width:"16",height:"7",rx:"3.5",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"6",cy:"8",r:"2",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("line",{x1:"7.5",y1:"9.5",x2:"9",y2:"11",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("rect",{x:"9.5",y:"7.5",width:"6",height:"1",rx:".5",fill:w,opacity:".12"})]});case"form":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"1.5",width:"5.5",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"2",y:"3.5",width:"16",height:"3",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2",y:"8",width:"7",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"2",y:"10",width:"16",height:"3",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"12",y:"14",width:"6",height:"2",rx:".75",stroke:w,strokeWidth:K})]});case"tabs":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"5",width:"18",height:"10",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"1",y:"2",width:"6",height:"3.5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2.5",y:"3.25",width:"3",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"7",y:"2",width:"6",height:"3.5",rx:".75",stroke:w,strokeWidth:K})]});case"dropdown":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"16",height:"4",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3.5",y:"3.5",width:"7",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("path",{d:"M15 3.5l1.5 1.5L18 3.5",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("rect",{x:"2",y:"7",width:"16",height:"7",rx:"1",stroke:w,strokeWidth:K,strokeDasharray:"2 1",opacity:".3"})]});case"toggle":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"4",y:"5",width:"12",height:"6",rx:"3",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"13",cy:"8",r:"2",fill:w,opacity:".3"})]});case"avatar":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"10",cy:"8",r:"6",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"10",cy:"6.5",r:"2",stroke:w,strokeWidth:K}),o.jsx("path",{d:"M6.5 13c0-2 1.5-3.5 3.5-3.5s3.5 1.5 3.5 3.5",stroke:w,strokeWidth:K})]});case"badge":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"5",width:"14",height:"6",rx:"3",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6",y:"7.5",width:"8",height:"1",rx:".5",fill:w,opacity:".25"})]});case"breadcrumb":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1.5",y:"7",width:"3.5",height:"1",rx:".5",fill:w,opacity:".3"}),o.jsx("path",{d:"M6.5 7l1 1-1 1",stroke:w,strokeWidth:K,opacity:".2"}),o.jsx("rect",{x:"9",y:"7",width:"3.5",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("path",{d:"M14 7l1 1-1 1",stroke:w,strokeWidth:K,opacity:".2"}),o.jsx("rect",{x:"16.5",y:"7",width:"2",height:"1",rx:".5",fill:w,opacity:".15"})]});case"pagination":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"5.5",width:"3.5",height:"5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6.5",y:"5.5",width:"3.5",height:"5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"11",y:"5.5",width:"3.5",height:"5",rx:"1",fill:w,opacity:".15",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"15.5",y:"5.5",width:"3.5",height:"5",rx:"1",stroke:w,strokeWidth:K})]});case"progress":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"7",width:"16",height:"2",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2",y:"7",width:"10",height:"2",rx:"1",fill:w,opacity:".2"})]});case"toast":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"4",width:"16",height:"8",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"5",cy:"8",r:"1.5",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("rect",{x:"8",y:"6.5",width:"7",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"8",y:"9",width:"5",height:"1",rx:".5",fill:w,opacity:".12"})]});case"tooltip":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"3",width:"14",height:"7",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5.5",y:"5.5",width:"9",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("path",{d:"M9 10l1 2.5 1-2.5",stroke:w,strokeWidth:K})]});case"pricing":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"1",width:"16",height:"14",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6",y:"3",width:"8",height:"1.5",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"7",y:"5.5",width:"6",height:"2",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"5",y:"9",width:"10",height:"1",rx:".5",fill:w,opacity:".1"}),o.jsx("rect",{x:"5",y:"11",width:"10",height:"1",rx:".5",fill:w,opacity:".1"}),o.jsx("rect",{x:"6",y:"13",width:"8",height:"1.5",rx:".5",fill:w,opacity:".2"})]});case"testimonial":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"1",width:"16",height:"14",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("text",{x:"4",y:"5.5",fontSize:"4",fill:w,opacity:".2",fontFamily:"serif",children:"“"}),o.jsx("rect",{x:"4",y:"7",width:"12",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"4",y:"9",width:"9",height:"1",rx:".5",fill:w,opacity:".12"}),o.jsx("circle",{cx:"5.5",cy:"12.5",r:"1.5",stroke:w,strokeWidth:K,opacity:".25"}),o.jsx("rect",{x:"8",y:"12",width:"5",height:"1",rx:".5",fill:w,opacity:".15"})]});case"cta":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"2",width:"18",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5",y:"4.5",width:"10",height:"1.5",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"6",y:"7.5",width:"8",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"7",y:"10",width:"6",height:"2.5",rx:"1",stroke:w,strokeWidth:K})]});case"alert":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"4",width:"16",height:"8",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"6",cy:"8",r:"2",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("line",{x1:"6",y1:"7",x2:"6",y2:"8.5",stroke:w,strokeWidth:"0.6",opacity:".5"}),o.jsx("circle",{cx:"6",cy:"9.3",r:".3",fill:w,opacity:".5"}),o.jsx("rect",{x:"9.5",y:"7",width:"6",height:"1",rx:".5",fill:w,opacity:".2"})]});case"banner":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1",y:"5",width:"18",height:"6",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"4",y:"7.5",width:"8",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"14",y:"7",width:"3.5",height:"2",rx:".75",stroke:w,strokeWidth:K})]});case"stat":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"2",width:"14",height:"12",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6",y:"4.5",width:"8",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"5",y:"7",width:"10",height:"2.5",rx:".5",fill:w,opacity:".3"}),o.jsx("rect",{x:"7",y:"11",width:"6",height:"1",rx:".5",fill:w,opacity:".12"})]});case"stepper":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"4",cy:"8",r:"2",fill:w,opacity:".2",stroke:w,strokeWidth:K}),o.jsx("line",{x1:"6",y1:"8",x2:"8",y2:"8",stroke:w,strokeWidth:".4",opacity:".3"}),o.jsx("circle",{cx:"10",cy:"8",r:"2",stroke:w,strokeWidth:K}),o.jsx("line",{x1:"12",y1:"8",x2:"14",y2:"8",stroke:w,strokeWidth:".4",opacity:".3"}),o.jsx("circle",{cx:"16",cy:"8",r:"2",stroke:w,strokeWidth:K})]});case"tag":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"5",width:"14",height:"6",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5.5",y:"7.5",width:"6",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("line",{x1:"14",y1:"6.5",x2:"15.5",y2:"9.5",stroke:w,strokeWidth:K,opacity:".2"}),o.jsx("line",{x1:"15.5",y1:"6.5",x2:"14",y2:"9.5",stroke:w,strokeWidth:K,opacity:".2"})]});case"rating":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("path",{d:"M4 5.5l1 2 2.2.3-1.6 1.5.4 2.2L4 10.3l-2 1.2.4-2.2L.8 7.8 3 7.5z",fill:w,opacity:".25"}),o.jsx("path",{d:"M10 5.5l1 2 2.2.3-1.6 1.5.4 2.2L10 10.3l-2 1.2.4-2.2L6.8 7.8 9 7.5z",fill:w,opacity:".25"}),o.jsx("path",{d:"M16 5.5l1 2 2.2.3-1.6 1.5.4 2.2L16 10.3l-2 1.2.4-2.2-1.6-1.5 2.2-.3z",stroke:w,strokeWidth:K,opacity:".25"})]});case"map":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"16",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("line",{x1:"2",y1:"6",x2:"18",y2:"10",stroke:w,strokeWidth:".3",opacity:".15"}),o.jsx("line",{x1:"7",y1:"2",x2:"11",y2:"14",stroke:w,strokeWidth:".3",opacity:".15"}),o.jsx("path",{d:"M10 5c-1.7 0-3 1.3-3 3 0 2.5 3 5 3 5s3-2.5 3-5c0-1.7-1.3-3-3-3z",fill:w,opacity:".15",stroke:w,strokeWidth:K})]});case"timeline":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("line",{x1:"5",y1:"2",x2:"5",y2:"14",stroke:w,strokeWidth:".4",opacity:".25"}),o.jsx("circle",{cx:"5",cy:"4",r:"1.5",fill:w,opacity:".2",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"8",y:"3",width:"8",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("circle",{cx:"5",cy:"8.5",r:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"8",y:"7.5",width:"6",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("circle",{cx:"5",cy:"13",r:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"8",y:"12",width:"7",height:"1",rx:".5",fill:w,opacity:".15"})]});case"fileUpload":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"2",width:"14",height:"12",rx:"1.5",stroke:w,strokeWidth:K,strokeDasharray:"2 1"}),o.jsx("path",{d:"M10 10V5.5m0 0L7.5 8m2.5-2.5L12.5 8",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("rect",{x:"7",y:"11.5",width:"6",height:"1",rx:".5",fill:w,opacity:".15"})]});case"codeBlock":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"16",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"4",cy:"4",r:".6",fill:w,opacity:".3"}),o.jsx("circle",{cx:"5.5",cy:"4",r:".6",fill:w,opacity:".3"}),o.jsx("circle",{cx:"7",cy:"4",r:".6",fill:w,opacity:".3"}),o.jsx("rect",{x:"4",y:"7",width:"7",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("rect",{x:"6",y:"9",width:"5",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"4",y:"11",width:"8",height:"1",rx:".5",fill:w,opacity:".12"})]});case"calendar":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"3",width:"16",height:"12",rx:"1",stroke:w,strokeWidth:K}),o.jsx("line",{x1:"2",y1:"6.5",x2:"18",y2:"6.5",stroke:w,strokeWidth:".4",opacity:".25"}),o.jsx("rect",{x:"5",y:"4",width:"1",height:"1.5",rx:".3",fill:w,opacity:".2"}),o.jsx("rect",{x:"14",y:"4",width:"1",height:"1.5",rx:".3",fill:w,opacity:".2"}),o.jsx("circle",{cx:"7",cy:"9",r:".6",fill:w,opacity:".2"}),o.jsx("circle",{cx:"10",cy:"9",r:".6",fill:w,opacity:".2"}),o.jsx("circle",{cx:"13",cy:"9",r:".6",fill:w,opacity:".3"}),o.jsx("circle",{cx:"7",cy:"12",r:".6",fill:w,opacity:".2"}),o.jsx("circle",{cx:"10",cy:"12",r:".6",fill:w,opacity:".2"})]});case"notification":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"3",width:"16",height:"10",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"5.5",cy:"8",r:"2",stroke:w,strokeWidth:K,opacity:".25"}),o.jsx("rect",{x:"9",y:"6",width:"6",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"9",y:"8.5",width:"4.5",height:"1",rx:".5",fill:w,opacity:".12"}),o.jsx("circle",{cx:"16.5",cy:"4.5",r:"1.5",fill:w,opacity:".25"})]});case"productCard":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"1",width:"14",height:"14",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3",y:"1",width:"14",height:"6",rx:"1",fill:w,opacity:".04"}),o.jsx("rect",{x:"5",y:"8.5",width:"7",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"5",y:"10.5",width:"4",height:"1.5",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"12",y:"12",width:"4",height:"2",rx:".75",stroke:w,strokeWidth:K})]});case"profile":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"10",cy:"5",r:"3",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5",y:"10",width:"10",height:"1.5",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"7",y:"12.5",width:"6",height:"1",rx:".5",fill:w,opacity:".12"})]});case"drawer":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"9",y:"1",width:"10",height:"14",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"10.5",y:"4",width:"5",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"10.5",y:"6.5",width:"7",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"10.5",y:"9",width:"6",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"1",y:"1",width:"7",height:"14",rx:"1",stroke:w,strokeWidth:K,opacity:".15"})]});case"popover":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"2",width:"14",height:"9",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5",y:"4.5",width:"8",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"5",y:"7",width:"6",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("path",{d:"M9 11l1 2.5 1-2.5",stroke:w,strokeWidth:K})]});case"logo":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"3",width:"10",height:"10",rx:"2",stroke:w,strokeWidth:K}),o.jsx("path",{d:"M5 9.5l2-4 2 4",stroke:w,strokeWidth:K,opacity:".3"}),o.jsx("rect",{x:"14",y:"6",width:"4",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("rect",{x:"14",y:"8.5",width:"3",height:"1",rx:".5",fill:w,opacity:".12"})]});case"faq":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("text",{x:"2.5",y:"5.5",fontSize:"4",fill:w,opacity:".3",fontWeight:"bold",children:"?"}),o.jsx("rect",{x:"7",y:"3",width:"10",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"7",y:"5.5",width:"8",height:"1",rx:".5",fill:w,opacity:".12"}),o.jsx("text",{x:"2.5",y:"11.5",fontSize:"4",fill:w,opacity:".3",fontWeight:"bold",children:"?"}),o.jsx("rect",{x:"7",y:"9",width:"9",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"7",y:"11.5",width:"7",height:"1",rx:".5",fill:w,opacity:".12"})]});case"gallery":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1.5",y:"1.5",width:"5",height:"5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"7.5",y:"1.5",width:"5",height:"5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"13.5",y:"1.5",width:"5",height:"5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"1.5",y:"9.5",width:"5",height:"5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"7.5",y:"9.5",width:"5",height:"5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"13.5",y:"9.5",width:"5",height:"5",rx:".75",stroke:w,strokeWidth:K})]});case"checkbox":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"5",y:"4",width:"8",height:"8",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("path",{d:"M7.5 8l1.5 1.5 3-3",stroke:w,strokeWidth:K,opacity:".35"})]});case"radio":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"10",cy:"8",r:"4",stroke:w,strokeWidth:K}),o.jsx("circle",{cx:"10",cy:"8",r:"2",fill:w,opacity:".3"})]});case"slider":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"7.5",width:"16",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"2",y:"7.5",width:"10",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("circle",{cx:"12",cy:"8",r:"2.5",stroke:w,strokeWidth:K})]});case"datePicker":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"1",width:"16",height:"5",rx:"1",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"3.5",y:"3",width:"5",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("rect",{x:"14",y:"2.5",width:"2.5",height:"2",rx:".5",fill:w,opacity:".12"}),o.jsx("rect",{x:"2",y:"7",width:"16",height:"8",rx:"1",stroke:w,strokeWidth:K,strokeDasharray:"2 1",opacity:".3"}),o.jsx("circle",{cx:"6",cy:"10",r:".6",fill:w,opacity:".2"}),o.jsx("circle",{cx:"10",cy:"10",r:".6",fill:w,opacity:".3"}),o.jsx("circle",{cx:"14",cy:"10",r:".6",fill:w,opacity:".2"}),o.jsx("circle",{cx:"6",cy:"13",r:".6",fill:w,opacity:".2"}),o.jsx("circle",{cx:"10",cy:"13",r:".6",fill:w,opacity:".2"})]});case"skeleton":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"16",height:"3",rx:"1",fill:w,opacity:".08"}),o.jsx("rect",{x:"2",y:"7",width:"10",height:"2",rx:".75",fill:w,opacity:".08"}),o.jsx("rect",{x:"2",y:"11",width:"13",height:"2",rx:".75",fill:w,opacity:".08"})]});case"chip":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"1.5",y:"5",width:"10",height:"6",rx:"3",fill:w,opacity:".08",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"4",y:"7.5",width:"4",height:"1",rx:".5",fill:w,opacity:".25"}),o.jsx("line",{x1:"9.5",y1:"6.5",x2:"10.5",y2:"9.5",stroke:w,strokeWidth:K,opacity:".2"}),o.jsx("line",{x1:"10.5",y1:"6.5",x2:"9.5",y2:"9.5",stroke:w,strokeWidth:K,opacity:".2"}),o.jsx("rect",{x:"13",y:"5",width:"5.5",height:"6",rx:"3",stroke:w,strokeWidth:K,opacity:".25"})]});case"icon":return o.jsx("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:o.jsx("path",{d:"M10 3l1.5 3 3.5.5-2.5 2.5.5 3.5L10 11l-3 1.5.5-3.5L5 6.5l3.5-.5z",stroke:w,strokeWidth:K,opacity:".3"})});case"spinner":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"10",cy:"8",r:"5",stroke:w,strokeWidth:K,opacity:".12"}),o.jsx("path",{d:"M10 3a5 5 0 0 1 5 5",stroke:w,strokeWidth:K,opacity:".35",strokeLinecap:"round"})]});case"feature":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"2",width:"5",height:"5",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("path",{d:"M4.5 3.5v3m-1.5-1.5h3",stroke:w,strokeWidth:K,opacity:".25"}),o.jsx("rect",{x:"9",y:"2.5",width:"8",height:"1.5",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"9",y:"5.5",width:"6",height:"1",rx:".5",fill:w,opacity:".12"}),o.jsx("rect",{x:"2",y:"10",width:"5",height:"5",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"9",y:"10.5",width:"7",height:"1.5",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"9",y:"13.5",width:"5",height:"1",rx:".5",fill:w,opacity:".12"})]});case"team":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("circle",{cx:"5",cy:"5",r:"2.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"2.5",y:"9",width:"5",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("circle",{cx:"15",cy:"5",r:"2.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"12.5",y:"9",width:"5",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("circle",{cx:"10",cy:"5",r:"2.5",stroke:w,strokeWidth:K,opacity:".5"}),o.jsx("rect",{x:"7.5",y:"9",width:"5",height:"1",rx:".5",fill:w,opacity:".15"}),o.jsx("rect",{x:"4",y:"12",width:"12",height:"1",rx:".5",fill:w,opacity:".1"})]});case"login":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"3",y:"1",width:"14",height:"14",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6",y:"3",width:"8",height:"1.5",rx:".5",fill:w,opacity:".25"}),o.jsx("rect",{x:"5",y:"5.5",width:"10",height:"3",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"5",y:"9.5",width:"10",height:"3",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"6.5",y:"13.5",width:"7",height:"2",rx:".75",fill:w,opacity:".2"})]});case"contact":return o.jsxs("svg",{viewBox:"0 0 20 16",width:"20",height:"16",fill:"none",children:[o.jsx("rect",{x:"2",y:"1",width:"16",height:"14",rx:"1.5",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"4",y:"3",width:"5",height:"1",rx:".5",fill:w,opacity:".2"}),o.jsx("rect",{x:"4",y:"5",width:"12",height:"2.5",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"4",y:"8.5",width:"12",height:"4",rx:".75",stroke:w,strokeWidth:K}),o.jsx("rect",{x:"11",y:"13.5",width:"5",height:"1.5",rx:".5",fill:w,opacity:".2"})]});default:return null}}function K5({activeType:s,onSelect:r,onDragStart:c,scrollRef:u,fadeClass:f,blankCanvas:d}){return o.jsx("div",{ref:u,className:`${ee.placeScroll} ${f||""}`,children:ey.map(m=>o.jsxs("div",{className:ee.paletteSection,children:[o.jsx("div",{className:ee.paletteSectionTitle,children:m.section}),m.items.map(g=>o.jsxs("button",{type:"button","aria-pressed":s===g.type,className:`${ee.paletteItem} ${s===g.type?ee.active:""} ${d?ee.wireframe:""}`,onClick:()=>r(g.type),onMouseDown:p=>{p.button===0&&c(g.type,p)},children:[o.jsx("span",{className:ee.paletteItemIcon,"aria-hidden":"true",children:o.jsx(W5,{type:g.type})}),o.jsx("span",{className:ee.paletteItemLabel,children:g.label})]},g.type))]},m.section))})}function Z5({value:s,suffix:r}){const[c,u]=h.useState(null),[f,d]=h.useState(r),[m,g]=h.useState("up"),p=h.useRef(s),j=h.useRef(r),k=h.useRef(),b=c!==null&&f!==r;return h.useEffect(()=>{if(s!==p.current){if(s===0){p.current=s,j.current=r,u(null);return}g(s>p.current?"up":"down"),u(p.current),d(j.current),p.current=s,j.current=r,clearTimeout(k.current),k.current=Ht(()=>u(null),250)}else j.current=r},[s,r]),c===null?o.jsxs(o.Fragment,{children:[s,r?` ${r}`:""]}):b?o.jsxs("span",{className:ee.rollingWrap,children:[o.jsxs("span",{style:{visibility:"hidden"},children:[s," ",r]}),o.jsxs("span",{className:`${ee.rollingNum} ${m==="up"?ee.exitUp:ee.exitDown}`,children:[c," ",f]},`o${c}-${s}`),o.jsxs("span",{className:`${ee.rollingNum} ${m==="up"?ee.enterUp:ee.enterDown}`,children:[s," ",r]},`n${s}`)]}):o.jsxs(o.Fragment,{children:[o.jsxs("span",{className:ee.rollingWrap,children:[o.jsx("span",{style:{visibility:"hidden"},children:s}),o.jsx("span",{className:`${ee.rollingNum} ${m==="up"?ee.exitUp:ee.exitDown}`,children:c},`o${c}-${s}`),o.jsx("span",{className:`${ee.rollingNum} ${m==="up"?ee.enterUp:ee.enterDown}`,children:s},`n${s}`)]}),r?` ${r}`:""]})}function J5({activeType:s,onSelect:r,isDarkMode:c,sectionCount:u,onDetectSections:f,visible:d,onExited:m,placementCount:g,onClearPlacements:p,onDragStart:j,blankCanvas:k,onBlankCanvasChange:b,wireframePurpose:v,onWireframePurposeChange:z,Tooltip:E}){const{ref:$,mounted:B}=ty(d,{onExited:m}),[F,Me]=h.useState(!1),[te,C]=h.useState(!0),ie=h.useRef(0),de=h.useRef(""),W=h.useRef(null),[Ee,ne]=h.useState(""),Ce=g>0||u>0,pe=g+u;return pe>0&&(ie.current=pe,de.current=k?pe===1?"Component":"Components":pe===1?"Change":"Changes"),h.useEffect(()=>{if(Ce)F?C(!1):(C(!0),Me(!0),nd(()=>{nd(()=>{C(!1)})}));else{C(!0);const V=Ht(()=>Me(!1),300);return()=>clearTimeout(V)}},[Ce]),h.useEffect(()=>{if(!d)return;const V=W.current;if(!V)return;const Ie=()=>ne(F5(V));V.addEventListener("scroll",Ie,{passive:!0});const he=new ResizeObserver(Ie);return he.observe(V),()=>{V.removeEventListener("scroll",Ie),he.disconnect()}},[d]),B?o.jsxs("div",{className:`${ee.palette} ${c?"":ee.light}`,ref:V=>{$.current=V,V==null||V.toggleAttribute("inert",!d)},"aria-hidden":!d,"data-feedback-toolbar":!0,"data-agentation-palette":!0,onClick:V=>V.stopPropagation(),onMouseDown:V=>V.stopPropagation(),children:[o.jsxs("div",{className:ee.paletteHeader,children:[o.jsx("div",{className:ee.paletteHeaderTitle,children:"Layout Mode"}),o.jsxs("div",{className:ee.paletteHeaderDesc,children:["Rearrange and resize existing elements, add new components, and explore layout ideas. Agent results may vary."," ",o.jsx("a",{href:"https://agentation.com/features#layout-mode",target:"_blank",rel:"noopener noreferrer",children:"Learn more."})]})]}),o.jsxs("button",{type:"button","aria-pressed":k,className:`${ee.canvasToggle} ${k?ee.active:""}`,onClick:()=>b(!k),children:[o.jsx("span",{className:ee.canvasToggleIcon,"aria-hidden":"true",children:o.jsxs("svg",{viewBox:"0 0 14 14",width:"14",height:"14",fill:"none",children:[o.jsx("rect",{x:"1",y:"1",width:"12",height:"12",rx:"2",stroke:"currentColor",strokeWidth:"1"}),o.jsx("circle",{cx:"4.5",cy:"4.5",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"7",cy:"4.5",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"9.5",cy:"4.5",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"4.5",cy:"7",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"7",cy:"7",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"9.5",cy:"7",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"4.5",cy:"9.5",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"7",cy:"9.5",r:"0.8",fill:"currentColor",opacity:".6"}),o.jsx("circle",{cx:"9.5",cy:"9.5",r:"0.8",fill:"currentColor",opacity:".6"})]})}),o.jsx("span",{className:ee.canvasToggleLabel,children:"Wireframe New Page"})]}),o.jsx("div",{className:`${ee.wireframePurposeWrap} ${k?"":ee.collapsed}`,"aria-hidden":!k,ref:V=>{V==null||V.toggleAttribute("inert",!k)},children:o.jsx("div",{className:ee.wireframePurposeInner,children:o.jsx("textarea",{className:ee.wireframePurposeInput,placeholder:"Describe this page to provide additional context for your agent.",value:v,onChange:V=>z(V.target.value),rows:2})})}),o.jsx(K5,{activeType:s,onSelect:r,onDragStart:j,scrollRef:W,fadeClass:Ee,blankCanvas:k}),F&&o.jsx("div",{className:`${ee.paletteFooterWrap} ${te?ee.footerHidden:""}`,children:o.jsx("div",{className:ee.paletteFooterInner,children:o.jsx("div",{className:ee.paletteFooterInnerContent,children:o.jsxs("div",{className:ee.paletteFooter,children:[o.jsx("span",{className:ee.paletteFooterCount,children:o.jsx(Z5,{value:ie.current,suffix:de.current})}),o.jsx("button",{className:ee.paletteFooterClear,onClick:p,children:"Clear"})]})})})})]}):null}var P5=new Set(["nav","header","main","section","article","footer","aside"]),R_={banner:"Header",navigation:"Navigation",main:"Main Content",contentinfo:"Footer",complementary:"Sidebar",region:"Section"},I0={nav:"Navigation",header:"Header",main:"Main Content",section:"Section",article:"Article",footer:"Footer",aside:"Sidebar"},e2=new Set(["script","style","noscript","link","meta"]),t2=40;function ny(s){let r=s;for(;r&&r!==document.body&&r!==document.documentElement;){const c=window.getComputedStyle(r).position;if(c==="fixed"||c==="sticky")return!0;r=r.parentElement}return!1}function Ba(s){const r=s.tagName.toLowerCase();if(["nav","header","footer","main"].includes(r)&&document.querySelectorAll(r).length===1)return r;if(s.id)return`#${CSS.escape(s.id)}`;if(s.className&&typeof s.className=="string"){const f=s.className.split(/\s+/).filter(d=>d.length>0).find(d=>d.length>2&&!/^[a-zA-Z0-9]{6,}$/.test(d)&&!/^[a-z]{1,2}$/.test(d));if(f){const d=`${r}.${CSS.escape(f)}`;if(document.querySelectorAll(d).length===1)return d}}const c=s.parentElement;if(c){const f=Array.from(c.children).indexOf(s)+1;return`${c===document.body?"body":Ba(c)} > ${r}:nth-child(${f})`}return r}function ld(s){var m;const r=s.tagName.toLowerCase(),c=s.getAttribute("aria-label");if(c)return c;const u=s.getAttribute("role");if(u&&R_[u])return R_[u];if(I0[r])return I0[r];const f=s.querySelector("h1, h2, h3, h4, h5, h6");if(f){const g=(m=f.textContent)==null?void 0:m.trim();if(g&&g.length<=50)return g;if(g)return g.slice(0,47)+"..."}const{name:d}=qr(s);return d.charAt(0).toUpperCase()+d.slice(1)}function ly(s){const r=s.className;return typeof r!="string"||!r?null:r.split(/\s+/).map(u=>u.replace(/[_][a-zA-Z0-9]{5,}.*$/,"")).find(u=>u.length>2&&!/^[a-z]{1,2}$/.test(u))||null}function oy(s){var u;const r=(u=s.textContent)==null?void 0:u.trim();if(!r)return null;const c=r.replace(/\s+/g," ");return c.length<=30?c:c.slice(0,30)+"…"}function n2(){const s=document.querySelector("main")||document.body,r=Array.from(s.children);let c=r;s!==document.body&&r.length<3&&(c=Array.from(document.body.children));const u=[];return c.forEach((f,d)=>{if(!(f instanceof HTMLElement))return;const m=f.tagName.toLowerCase();if(e2.has(m)||f.hasAttribute("data-feedback-toolbar")||f.closest("[data-feedback-toolbar]"))return;const g=window.getComputedStyle(f);if(g.display==="none"||g.visibility==="hidden")return;const p=f.getBoundingClientRect();if(p.height<t2)return;const j=P5.has(m),k=f.getAttribute("role")&&R_[f.getAttribute("role")],b=m==="div"&&p.height>=60;if(!j&&!k&&!b)return;const v=window.scrollY,z=ny(f),E={x:p.x,y:z?p.y:p.y+v,width:p.width,height:p.height};u.push({id:`rs-${Date.now()}-${Math.random().toString(36).slice(2,7)}`,label:ld(f),tagName:m,selector:Ba(f),role:f.getAttribute("role"),className:ly(f),textSnippet:oy(f),originalRect:E,currentRect:{...E},originalIndex:d,isFixed:z})}),u}function l2(s){const r=window.scrollY,c=s.getBoundingClientRect(),u=ny(s),f={x:c.x,y:u?c.y:c.y+r,width:c.width,height:c.height},d=s.parentElement;let m=0;return d&&(m=Array.from(d.children).indexOf(s)),{id:`rs-${Date.now()}-${Math.random().toString(36).slice(2,7)}`,label:ld(s),tagName:s.tagName.toLowerCase(),selector:Ba(s),role:s.getAttribute("role"),className:ly(s),textSnippet:oy(s),originalRect:f,currentRect:{...f},originalIndex:m,isFixed:u}}var X0={bg:"rgba(59, 130, 246, 0.08)",border:"rgba(59, 130, 246, 0.5)",pill:"#3b82f6"},Q0=["nw","n","ne","e","se","s","sw","w"],Iu=24,G0=16,Xu=5;function V0(s,r,c,u){let f=1/0,d=1/0;const m=s.x,g=s.x+s.width,p=s.x+s.width/2,j=s.y,k=s.y+s.height,b=s.y+s.height/2,v=[];for(const W of r)c.has(W.id)||v.push(W.currentRect);u&&v.push(...u);for(const W of v){const Ee=W.x,ne=W.x+W.width,Ce=W.x+W.width/2,pe=W.y,V=W.y+W.height,Ie=W.y+W.height/2;for(const he of[m,g,p])for(const ue of[Ee,ne,Ce]){const G=ue-he;Math.abs(G)<Xu&&Math.abs(G)<Math.abs(f)&&(f=G)}for(const he of[j,k,b])for(const ue of[pe,V,Ie]){const G=ue-he;Math.abs(G)<Xu&&Math.abs(G)<Math.abs(d)&&(d=G)}}const z=Math.abs(f)<Xu?f:0,E=Math.abs(d)<Xu?d:0,$=[],B=new Set,F=m+z,Me=g+z,te=p+z,C=j+E,ie=k+E,de=b+E;for(const W of v){const Ee=W.x,ne=W.x+W.width,Ce=W.x+W.width/2,pe=W.y,V=W.y+W.height,Ie=W.y+W.height/2;for(const he of[Ee,Ce,ne])for(const ue of[F,te,Me])if(Math.abs(ue-he)<.5){const G=`x:${Math.round(he)}`;B.has(G)||(B.add(G),$.push({axis:"x",pos:he}))}for(const he of[pe,Ie,V])for(const ue of[C,de,ie])if(Math.abs(ue-he)<.5){const G=`y:${Math.round(he)}`;B.has(G)||(B.add(G),$.push({axis:"y",pos:he}))}}return{dx:z,dy:E,guides:$}}var o2=new Set(["script","style","noscript","link","meta","br","hr"]);function F0(s){let r=s;for(;r&&r!==document.body&&r!==document.documentElement;){if(r.closest("[data-feedback-toolbar]"))return null;if(o2.has(r.tagName.toLowerCase())){r=r.parentElement;continue}const c=r.getBoundingClientRect();if(c.width>=G0&&c.height>=G0)return r;r=r.parentElement}return null}function s2({rearrangeState:s,onChange:r,isDarkMode:c,exiting:u,className:f,blankCanvas:d,extraSnapRects:m,onSelectionChange:g,deselectSignal:p,onDragMove:j,onDragEnd:k,clearing:b}){const{sections:v}=s,z=h.useRef(s);z.current=s;const[E,$]=h.useState(new Set);h.useEffect(()=>{b&&$(new Set)},[b]);const B=h.useRef(p);h.useEffect(()=>{p!==B.current&&(B.current=p,$(new Set))},[p]);const[F,Me]=h.useState(null),[te,C]=h.useState(!1),ie=h.useRef(!1),de=h.useCallback(O=>{const I=v.find(P=>P.id===O);I&&(ie.current=!!I.note,Me(O),C(!1))},[v]),W=h.useCallback(()=>{F&&(C(!0),Ht(()=>{Me(null),C(!1)},150))},[F]),Ee=h.useCallback(O=>{F&&(r({...s,sections:v.map(I=>I.id===F?{...I,note:O.trim()||void 0}:I)}),W())},[F,v,s,r,W]);h.useEffect(()=>{u&&F&&W()},[u]);const[ne,Ce]=h.useState(new Set),pe=h.useRef(new Map),[V,Ie]=h.useState(null),[he,ue]=h.useState(null),[G,we]=h.useState([]),[fe,Oe]=h.useState(0),Se=h.useRef(null),Qe=h.useRef(new Set),Ut=h.useRef(new Map),[Ye,Ne]=h.useState(new Map),[Xt,xe]=h.useState(new Map),We=h.useRef(new Set),gn=h.useRef(new Map),Qt=h.useRef(g);Qt.current=g;const mn=h.useRef(j);mn.current=j;const Re=h.useRef(k);Re.current=k,h.useEffect(()=>{d&&$(new Set)},[d]);const[Ue,un]=h.useState(()=>!s.sections.some(O=>{const I=O.originalRect,P=O.currentRect;return Math.abs(I.x-P.x)>1||Math.abs(I.y-P.y)>1||Math.abs(I.width-P.width)>1||Math.abs(I.height-P.height)>1}));h.useEffect(()=>{if(!Ue){const O=Ht(()=>un(!0),380);return()=>clearTimeout(O)}},[]);const ul=h.useRef(new Set);h.useEffect(()=>{ul.current=new Set(v.map(O=>O.selector))},[v]),h.useEffect(()=>{const O=()=>Oe(window.scrollY);return O(),window.addEventListener("scroll",O,{passive:!0}),window.addEventListener("resize",O,{passive:!0}),()=>{window.removeEventListener("scroll",O),window.removeEventListener("resize",O)}},[]),h.useEffect(()=>{const O=I=>{if(Se.current){Ie(null);return}const P=document.elementFromPoint(I.clientX,I.clientY);if(!P){Ie(null);return}if(P.closest("[data-feedback-toolbar]")){Ie(null);return}if(P.closest("[data-design-placement]")){Ie(null);return}if(P.closest("[data-annotation-popup]")){Ie(null);return}const ae=F0(P);if(!ae){Ie(null);return}for(const ke of ul.current)try{const _e=document.querySelector(ke);if(_e&&(_e===ae||ae.contains(_e))){Ie(null);return}}catch{}const Be=ae.getBoundingClientRect();Ie({x:Be.x,y:Be.y,w:Be.width,h:Be.height})};return document.addEventListener("mousemove",O,{passive:!0}),()=>document.removeEventListener("mousemove",O)},[v]),h.useEffect(()=>{const O=document.body.style.userSelect;return document.body.style.webkitUserSelect="none",document.body.style.userSelect="none",()=>{document.body.style.webkitUserSelect=O,document.body.style.userSelect=O}},[]),h.useEffect(()=>{const O=I=>{var _e,at,ot,ct;if(Se.current||I.button!==0)return;const P=I.composedPath()[0]??I.target;if(!P||P.closest("[data-feedback-toolbar]")||P.closest("[data-design-placement]")||P.closest("[data-annotation-popup]"))return;const ae=F0(P);let Be=!1;if(ae)for(const Ke of ul.current)try{const Ze=document.querySelector(Ke);if(Ze&&(Ze===ae||ae.contains(Ze))){Be=!0;break}}catch{}const ke=!!(I.shiftKey||I.metaKey||I.ctrlKey);if(ae&&!Be){I.preventDefault(),I.stopPropagation();const Ke=l2(ae),Ze=[...v,Ke],bt=[...s.originalOrder,Ke.id];r({...s,sections:Ze,originalOrder:bt});const Bt=new Set([Ke.id]);$(Bt),(_e=Qt.current)==null||_e.call(Qt,Bt,ke),Ie(null);const pn=I.clientX,pt=I.clientY,At={x:Ke.currentRect.x,y:Ke.currentRect.y};Ke.originalRect;let kt=!1,ht=0,Nn=0;Se.current="move";const vt=je=>{var Mn;const Gt=je.clientX-pn,vn=je.clientY-pt;if(!kt&&(Math.abs(Gt)>2||Math.abs(vn)>2)&&(kt=!0),!kt)return;const ho={x:At.x+Gt,y:At.y+vn,width:Ke.currentRect.width,height:Ke.currentRect.height},Ql=V0(ho,Ze,new Set([Ke.id]),m);we(Ql.guides);const yn=Gt+Ql.dx,Fn=vn+Ql.dy;ht=yn,Nn=Fn;const Hn=oe().querySelector(`[data-rearrange-section="${Ke.id}"]`);Hn&&(Hn.style.transform=`translate(${yn}px, ${Fn}px)`),Ne(new Map([[Ke.id,{x:At.x+yn,y:At.y+Fn,width:Ke.currentRect.width,height:Ke.currentRect.height}]])),(Mn=mn.current)==null||Mn.call(mn,yn,Fn)},Je=()=>{var Gt;window.removeEventListener("mousemove",vt),window.removeEventListener("mouseup",Je),Se.current=null,we([]),Ne(new Map);const je=oe().querySelector(`[data-rearrange-section="${Ke.id}"]`);je&&(je.style.transform=""),kt&&r({...s,sections:Ze.map(vn=>vn.id===Ke.id?{...vn,currentRect:{...vn.currentRect,x:Math.max(0,At.x+ht),y:Math.max(0,At.y+Nn)}}:vn),originalOrder:bt}),(Gt=Re.current)==null||Gt.call(Re,ht,Nn,kt)};window.addEventListener("mousemove",vt),window.addEventListener("mouseup",Je)}else if(Be&&ae){I.preventDefault();for(const Ke of v)try{const Ze=document.querySelector(Ke.selector);if(Ze&&Ze===ae){const bt=new Set([Ke.id]);$(bt),(at=Qt.current)==null||at.call(Qt,bt,ke);return}}catch{}ke||($(new Set),(ot=Qt.current)==null||ot.call(Qt,new Set,!1))}else ke||($(new Set),(ct=Qt.current)==null||ct.call(Qt,new Set,!1))};return document.addEventListener("mousedown",O,!0),()=>document.removeEventListener("mousedown",O,!0)},[v,s,r]),h.useEffect(()=>{const O=I=>{const P=I.composedPath()[0]||I.target;if(!(P.tagName==="INPUT"||P.tagName==="TEXTAREA"||P.isContentEditable)){if((I.key==="Backspace"||I.key==="Delete")&&E.size>0){I.preventDefault();const ae=new Set(E);Ce(Be=>{const ke=new Set(Be);for(const _e of ae)ke.add(_e);return ke}),$(new Set),Ht(()=>{const Be=z.current;r({...Be,sections:Be.sections.filter(ke=>!ae.has(ke.id)),originalOrder:Be.originalOrder.filter(ke=>!ae.has(ke))}),Ce(ke=>{const _e=new Set(ke);for(const at of ae)_e.delete(at);return _e})},180);return}if(["ArrowUp","ArrowDown","ArrowLeft","ArrowRight"].includes(I.key)&&E.size>0){I.preventDefault();const ae=I.shiftKey?20:1,Be=I.key==="ArrowLeft"?-ae:I.key==="ArrowRight"?ae:0,ke=I.key==="ArrowUp"?-ae:I.key==="ArrowDown"?ae:0;r({...s,sections:v.map(_e=>E.has(_e.id)?{..._e,currentRect:{..._e.currentRect,x:Math.max(0,_e.currentRect.x+Be),y:Math.max(0,_e.currentRect.y+ke)}}:_e)});return}I.key==="Escape"&&E.size>0&&$(new Set)}};return document.addEventListener("keydown",O),()=>document.removeEventListener("keydown",O)},[E,v,s,r]);const Cn=h.useCallback((O,I)=>{var pn;if(O.button!==0)return;const P=O.target;if(P.closest(`.${ee.handle}`)||P.closest(`.${ee.deleteButton}`))return;O.preventDefault(),O.stopPropagation();let ae;O.shiftKey||O.metaKey||O.ctrlKey?(ae=new Set(E),ae.has(I)?ae.delete(I):ae.add(I)):E.has(I)?ae=new Set(E):ae=new Set([I]),$(ae),(ae.size!==E.size||[...ae].some(pt=>!E.has(pt)))&&((pn=Qt.current)==null||pn.call(Qt,ae,!!(O.shiftKey||O.metaKey||O.ctrlKey)));const ke=O.clientX,_e=O.clientY,at=new Map;for(const pt of v)ae.has(pt.id)&&at.set(pt.id,{x:pt.currentRect.x,y:pt.currentRect.y});Se.current="move";let ot=!1,ct=0,Ke=0;const Ze=new Map;for(const pt of v)if(ae.has(pt.id)){const At=oe().querySelector(`[data-rearrange-section="${pt.id}"]`);Ze.set(pt.id,{outlineEl:At,curW:pt.currentRect.width,curH:pt.currentRect.height})}const bt=pt=>{var Ql;const At=pt.clientX-ke,kt=pt.clientY-_e;if(At===0&&kt===0)return;ot=!0;let ht=1/0,Nn=1/0,vt=-1/0,Je=-1/0;for(const[yn,{curW:Fn,curH:Hn}]of Ze){const Mn=at.get(yn);if(!Mn)continue;const dl=Mn.x+At,sl=Mn.y+kt;ht=Math.min(ht,dl),Nn=Math.min(Nn,sl),vt=Math.max(vt,dl+Fn),Je=Math.max(Je,sl+Hn)}const je=V0({x:ht,y:Nn,width:vt-ht,height:Je-Nn},v,ae,m),Gt=At+je.dx,vn=kt+je.dy;ct=Gt,Ke=vn,we(je.guides);for(const[,{outlineEl:yn}]of Ze)yn&&(yn.style.transform=`translate(${Gt}px, ${vn}px)`);const ho=new Map;for(const[yn,{curW:Fn,curH:Hn}]of Ze){const Mn=at.get(yn);if(Mn){const dl={x:Math.max(0,Mn.x+Gt),y:Math.max(0,Mn.y+vn),width:Fn,height:Hn};ho.set(yn,dl)}}Ne(ho),(Ql=mn.current)==null||Ql.call(mn,Gt,vn)},Bt=pt=>{var At,kt;window.removeEventListener("mousemove",bt),window.removeEventListener("mouseup",Bt),Se.current=null,we([]),Ne(new Map);for(const[,{outlineEl:ht}]of Ze)ht&&(ht.style.transform="");if(ot){const ht=pt.clientX-ke,Nn=pt.clientY-_e;if(Math.abs(ht)<5&&Math.abs(Nn)<5)r({...s,sections:v.map(vt=>{const Je=at.get(vt.id);return Je?{...vt,currentRect:{...vt.currentRect,x:Je.x,y:Je.y}}:vt})});else{r({...s,sections:v.map(vt=>{const Je=at.get(vt.id);return Je?{...vt,currentRect:{...vt.currentRect,x:Math.max(0,Je.x+ct),y:Math.max(0,Je.y+Ke)}}:vt})}),(At=Re.current)==null||At.call(Re,ct,Ke,!0);return}}(kt=Re.current)==null||kt.call(Re,0,0,!1)};window.addEventListener("mousemove",bt),window.addEventListener("mouseup",Bt)},[E,v,s,r]),bn=h.useCallback((O,I,P)=>{O.preventDefault(),O.stopPropagation();const ae=v.find(bt=>bt.id===I);if(!ae)return;$(new Set([I])),Se.current="resize";const Be=O.clientX,ke=O.clientY,_e={...ae.currentRect};ae.originalRect;const at=_e.width/_e.height;let ot={..._e};const ct=oe().querySelector(`[data-rearrange-section="${I}"]`),Ke=bt=>{const Bt=bt.clientX-Be,pn=bt.clientY-ke;let pt=_e.x,At=_e.y,kt=_e.width,ht=_e.height;if(P.includes("e")&&(kt=Math.max(Iu,_e.width+Bt)),P.includes("w")&&(kt=Math.max(Iu,_e.width-Bt),pt=_e.x+_e.width-kt),P.includes("s")&&(ht=Math.max(Iu,_e.height+pn)),P.includes("n")&&(ht=Math.max(Iu,_e.height-pn),At=_e.y+_e.height-ht),bt.shiftKey)if(P.length===2){const vt=Math.abs(kt-_e.width),Je=Math.abs(ht-_e.height);vt>Je?ht=kt/at:kt=ht*at,P.includes("w")&&(pt=_e.x+_e.width-kt),P.includes("n")&&(At=_e.y+_e.height-ht)}else P==="e"||P==="w"?ht=kt/at:kt=ht*at,P==="w"&&(pt=_e.x+_e.width-kt),P==="n"&&(At=_e.y+_e.height-ht);ot={x:pt,y:At,width:kt,height:ht},ct&&(ct.style.left=`${pt}px`,ct.style.top=`${At-fe}px`,ct.style.width=`${kt}px`,ct.style.height=`${ht}px`),ue({x:bt.clientX+12,y:bt.clientY+12,text:`${Math.round(kt)} × ${Math.round(ht)}`}),Ne(new Map([[I,ot]]))},Ze=()=>{window.removeEventListener("mousemove",Ke),window.removeEventListener("mouseup",Ze),ue(null),Se.current=null,Ne(new Map),r({...s,sections:v.map(bt=>bt.id===I?{...bt,currentRect:ot}:bt)})};window.addEventListener("mousemove",Ke),window.addEventListener("mouseup",Ze)},[v,s,r,fe]),re=h.useCallback(O=>{Ce(I=>{const P=new Set(I);return P.add(O),P}),$(I=>{const P=new Set(I);return P.delete(O),P}),Ht(()=>{const I=z.current;r({...I,sections:I.sections.filter(P=>P.id!==O),originalOrder:I.originalOrder.filter(P=>P!==O)}),Ce(P=>{const ae=new Set(P);return ae.delete(O),ae})},180)},[r]),ye=O=>{const I=O.originalRect,P=O.currentRect;return Math.abs(I.x-P.x)>1||Math.abs(I.y-P.y)>1||Math.abs(I.width-P.width)>1||Math.abs(I.height-P.height)>1},Ge=O=>{const I=O.originalRect,P=O.currentRect;return Math.abs(I.x-P.x)>1||Math.abs(I.y-P.y)>1},nt=O=>{const I=O.originalRect,P=O.currentRect;return Math.abs(I.width-P.width)>1||Math.abs(I.height-P.height)>1};for(const O of v)Ut.current.has(O.id)||(Ge(O)?Ut.current.set(O.id,"move"):nt(O)&&Ut.current.set(O.id,"resize"));for(const O of Ut.current.keys())v.some(I=>I.id===O)||Ut.current.delete(O);const lt=v.filter(O=>{try{if(ne.has(O.id)||E.has(O.id))return!0;const I=document.querySelector(O.selector);if(!I)return!1;const P=I.getBoundingClientRect(),ae=O.originalRect;return Math.abs(P.width-ae.width)+Math.abs(P.height-ae.height)<200}catch{return!1}}),rt=lt.filter(O=>ye(O)),it=lt.filter(O=>!ye(O)),Fe=new Set(rt.map(O=>O.id));for(const O of Qe.current)Fe.has(O)||Qe.current.delete(O);const Nt=[...Fe].sort().join(",");for(const O of rt)gn.current.set(O.id,{currentRect:O.currentRect,originalRect:O.originalRect,isFixed:O.isFixed});h.useEffect(()=>{const O=We.current;We.current=Fe;const I=new Map;for(const P of O)if(!Fe.has(P)){if(!v.some(Be=>Be.id===P))continue;const ae=gn.current.get(P);ae&&(I.set(P,{orig:ae.originalRect,target:ae.currentRect,isFixed:ae.isFixed}),gn.current.delete(P))}if(I.size>0){xe(ae=>{const Be=new Map(ae);for(const[ke,_e]of I)Be.set(ke,_e);return Be});const P=Ht(()=>{xe(ae=>{const Be=new Map(ae);for(const ke of I.keys())Be.delete(ke);return Be})},250);return()=>clearTimeout(P)}},[Nt,v]);const an=h.useRef(null),oe=()=>{var O;return((O=an.current)==null?void 0:O.getRootNode())??document};return o.jsxs(o.Fragment,{children:[o.jsxs("div",{ref:an,className:`${ee.rearrangeOverlay} ${c?"":ee.light} ${u?ee.overlayExiting:""}${f?` ${f}`:""}`,"data-feedback-toolbar":!0,children:[V&&o.jsx("div",{className:ee.hoverHighlight,style:{left:V.x,top:V.y,width:V.w,height:V.h}}),it.map(O=>{const I=O.currentRect,P=O.isFixed?I.y:I.y-fe,ae=X0,Be=E.has(O.id);return o.jsxs("div",{"data-rearrange-section":O.id,className:`${ee.sectionOutline} ${Be?ee.selected:""} ${b||u||ne.has(O.id)?ee.exiting:""}`,style:{left:I.x,top:P,width:I.width,height:I.height,borderColor:ae.border,backgroundColor:ae.bg,...Ue?{}:{opacity:0,animation:"none",transition:"none"}},onMouseDown:ke=>Cn(ke,O.id),onDoubleClick:()=>de(O.id),children:[o.jsx("span",{className:ee.sectionLabel,style:{backgroundColor:ae.pill},children:O.label}),o.jsx("span",{className:`${ee.sectionAnnotation} ${O.note?ee.annotationVisible:""}`,children:(O.note&&pe.current.set(O.id,O.note),O.note||pe.current.get(O.id)||"")}),o.jsxs("span",{className:ee.sectionDimensions,children:[Math.round(I.width)," × ",Math.round(I.height)]}),o.jsx("div",{className:ee.deleteButton,onMouseDown:ke=>ke.stopPropagation(),onClick:()=>re(O.id),children:"✕"}),Q0.map(ke=>o.jsx("div",{className:`${ee.handle} ${ee[`handle${ke.charAt(0).toUpperCase()}${ke.slice(1)}`]}`,onMouseDown:_e=>bn(_e,O.id,ke)},ke))]},O.id)}),rt.map(O=>{const I=O.currentRect,P=O.isFixed?I.y:I.y-fe,ae=E.has(O.id),Be=Ge(O),ke=nt(O);if(d&&!ae)return null;const at=!Qe.current.has(O.id);return at&&Qe.current.add(O.id),o.jsxs("div",{"data-rearrange-section":O.id,className:`${ee.ghostOutline} ${ae?ee.selected:""} ${b||u||ne.has(O.id)?ee.exiting:""}`,style:{left:I.x,top:P,width:I.width,height:I.height,...Ue?{}:{opacity:0,animation:"none",transition:"none"},...at?{}:{animation:"none"}},onMouseDown:ot=>Cn(ot,O.id),onDoubleClick:()=>de(O.id),children:[o.jsx("span",{className:ee.sectionLabel,style:{backgroundColor:X0.pill},children:O.label}),o.jsx("span",{className:`${ee.sectionAnnotation} ${O.note?ee.annotationVisible:""}`,children:(O.note&&pe.current.set(O.id,O.note),O.note||pe.current.get(O.id)||"")}),o.jsxs("span",{className:ee.sectionDimensions,children:[Math.round(I.width)," × ",Math.round(I.height)]}),o.jsx("div",{className:ee.deleteButton,onMouseDown:ot=>ot.stopPropagation(),onClick:()=>re(O.id),children:"✕"}),Q0.map(ot=>o.jsx("div",{className:`${ee.handle} ${ee[`handle${ot.charAt(0).toUpperCase()}${ot.slice(1)}`]}`,onMouseDown:ct=>bn(ct,O.id,ot)},ot)),o.jsx("span",{className:ee.ghostBadge,children:(()=>{const ot=Ut.current.get(O.id);if(Be&&ke){const[ct,Ke]=ot==="resize"?["Resize","Move"]:["Move","Resize"];return o.jsxs(o.Fragment,{children:["Suggested ",ct," ",o.jsxs("span",{className:ee.ghostBadgeExtra,children:["& ",Ke]})]})}return`Suggested ${ke?"Resize":"Move"}`})()})]},O.id)})]}),!d&&(()=>{const O=[];for(const I of rt){const P=Ye.get(I.id);O.push({id:I.id,orig:I.originalRect,target:P||I.currentRect,isFixed:I.isFixed,isSelected:E.has(I.id),isExiting:ne.has(I.id)})}for(const[I,P]of Ye)if(!O.some(ae=>ae.id===I)){const ae=v.find(Be=>Be.id===I);ae&&O.push({id:I,orig:ae.originalRect,target:P,isFixed:ae.isFixed,isSelected:E.has(I)})}for(const[I,P]of Xt)O.some(ae=>ae.id===I)||O.push({id:I,orig:P.orig,target:P.target,isFixed:P.isFixed,isSelected:!1,isExiting:!0});return O.length===0?null:o.jsxs("svg",{className:`${ee.connectorSvg} ${b||u?ee.connectorExiting:""}`,children:[O.map(({id:I,orig:P,target:ae,isFixed:Be,isSelected:ke,isExiting:_e})=>{const at=P.x+P.width/2,ot=(Be?P.y:P.y-fe)+P.height/2,ct=ae.x+ae.width/2,Ke=(Be?ae.y:ae.y-fe)+ae.height/2,Ze=ct-at,bt=Ke-ot,Bt=Math.sqrt(Ze*Ze+bt*bt);if(Bt<2)return null;const pn=Math.min(1,Bt/40),pt=Math.min(Bt*.3,60),At=Bt>0?-bt/Bt:0,kt=Bt>0?Ze/Bt:0,ht=(at+ct)/2+At*pt,Nn=(ot+Ke)/2+kt*pt,vt=Ye.has(I),Je=vt||ke?1:.4,je=vt||ke?1:.5;return o.jsxs("g",{className:_e?ee.connectorExiting:"",children:[o.jsx("path",{className:ee.connectorLine,d:`M ${at} ${ot} Q ${ht} ${Nn} ${ct} ${Ke}`,fill:"none",stroke:"rgba(59, 130, 246, 0.45)",strokeWidth:"1.5",opacity:Je*pn}),o.jsx("circle",{className:ee.connectorDot,cx:at,cy:ot,r:4*pn,fill:"rgba(59, 130, 246, 0.8)",stroke:"#fff",strokeWidth:"1.5",opacity:je*pn,filter:"url(#connDotShadow)"}),o.jsx("circle",{className:ee.connectorDot,cx:ct,cy:Ke,r:4*pn,fill:"rgba(59, 130, 246, 0.8)",stroke:"#fff",strokeWidth:"1.5",opacity:je*pn,filter:"url(#connDotShadow)"})]},`conn-${I}`)}),o.jsx("defs",{children:o.jsx("filter",{id:"connDotShadow",x:"-50%",y:"-50%",width:"200%",height:"200%",children:o.jsx("feDropShadow",{dx:"0",dy:"0.5",stdDeviation:"1",floodOpacity:"0.15"})})})]})})(),F&&(()=>{const O=v.find(Ke=>Ke.id===F);if(!O)return null;const I=O.currentRect,P=O.isFixed?I.y:I.y-fe,ae=I.x+I.width/2,Be=P-8,ke=P+I.height+8,_e=Be>200,at=ke<window.innerHeight-100,ot=Math.max(160,Math.min(window.innerWidth-160,ae));let ct;return _e?ct={left:ot,bottom:window.innerHeight-Be}:at?ct={left:ot,top:ke}:ct={left:ot,top:Math.max(80,window.innerHeight/2-80)},o.jsx(K_,{element:O.label,placeholder:"Add a note about this section",initialValue:O.note??"",submitLabel:ie.current?"Save":"Set",onSubmit:Ee,onCancel:W,onDelete:ie.current?()=>{Ee("")}:void 0,isExiting:te,lightMode:!c,style:ct})})(),he&&o.jsx("div",{className:ee.sizeIndicator,style:{left:he.x,top:he.y},"data-feedback-toolbar":!0,children:he.text}),G.map((O,I)=>o.jsx("div",{className:ee.guideLine,style:O.axis==="x"?{position:"fixed",left:O.pos,top:0,width:1,height:"100vh"}:{position:"fixed",left:0,top:O.pos-fe,width:"100vw",height:1}},`${O.axis}-${O.pos}-${I}`))]})}var z_=new Set(["script","style","noscript","link","meta","br","hr"]);function a2(){const s=document.querySelector("main")||document.body,r=[],c=Array.from(s.children),u=s!==document.body&&c.length<3?Array.from(document.body.children):c;for(const f of u){if(!(f instanceof HTMLElement)||z_.has(f.tagName.toLowerCase())||f.hasAttribute("data-feedback-toolbar"))continue;const d=window.getComputedStyle(f);if(d.display==="none"||d.visibility==="hidden")continue;const m=f.getBoundingClientRect();if(!(m.height<10||m.width<10)){r.push({label:ld(f),selector:Ba(f),top:m.top,bottom:m.bottom,left:m.left,right:m.right,area:m.width*m.height});for(const g of Array.from(f.children)){if(!(g instanceof HTMLElement)||z_.has(g.tagName.toLowerCase())||g.hasAttribute("data-feedback-toolbar"))continue;const p=window.getComputedStyle(g);if(p.display==="none"||p.visibility==="hidden")continue;const j=g.getBoundingClientRect();j.height<10||j.width<10||r.push({label:ld(g),selector:Ba(g),top:j.top,bottom:j.bottom,left:j.left,right:j.right,area:j.width*j.height})}}}return r}function r2(s){const r=window.scrollY;return s.map(({label:c,selector:u,rect:f})=>{const d=f.y-r;return{label:c,selector:u,top:d,bottom:d+f.height,left:f.x,right:f.x+f.width,area:f.width*f.height}})}function i2(s){const r=window.scrollY,c=s.y-r,u=s.x;return{top:c,bottom:c+s.height,left:u,right:u+s.width,area:s.width*s.height}}function A_(s,r){const c=r?r2(r):a2(),u=i2(s);let f=null,d=null,m=null,g=null,p=null;for(const E of c){if(Math.abs(E.left-u.left)<2&&Math.abs(E.top-u.top)<2&&Math.abs(E.right-E.left-s.width)<2&&Math.abs(E.bottom-E.top-s.height)<2)continue;E.left<=u.left+2&&E.right>=u.right-2&&E.top<=u.top+2&&E.bottom>=u.bottom-2&&E.area>u.area*1.5&&(!p||E.area<p._area)&&(p={label:E.label,selector:E.selector,_area:E.area});const $=u.right>E.left+5&&u.left<E.right-5,B=u.bottom>E.top+5&&u.top<E.bottom-5;if($&&E.bottom<=u.top+5){const F=Math.round(u.top-E.bottom);(!f||F<f._dist)&&(f={label:E.label,selector:E.selector,gap:Math.max(0,F),_dist:F})}if($&&E.top>=u.bottom-5){const F=Math.round(E.top-u.bottom);(!d||F<d._dist)&&(d={label:E.label,selector:E.selector,gap:Math.max(0,F),_dist:F})}if(B&&E.right<=u.left+5){const F=Math.round(u.left-E.right);(!m||F<m._dist)&&(m={label:E.label,selector:E.selector,gap:Math.max(0,F),_dist:F})}if(B&&E.left>=u.right-5){const F=Math.round(E.left-u.right);(!g||F<g._dist)&&(g={label:E.label,selector:E.selector,gap:Math.max(0,F),_dist:F})}}const j=window.innerWidth,k=window.innerHeight,b=u2(s,j),v=E=>E?{label:E.label,selector:E.selector,gap:E.gap}:null,z=c2(u,s,j,k,p?{label:p.label,selector:p.selector,_area:p._area}:null,c);return{above:v(f),below:v(d),left:v(m),right:v(g),alignment:b,containedIn:p?{label:p.label,selector:p.selector}:null,outOfBounds:z}}function c2(s,r,c,u,f,d){const m={};let g=!1;const p=[];if(s.left<-2&&p.push("left"),s.right>c+2&&p.push("right"),s.top<-2&&p.push("top"),s.bottom>u+2&&p.push("bottom"),p.length>0&&(m.viewport=p,g=!0),f){const j=d.find(k=>k.label===f.label&&k.selector===f.selector&&Math.abs(k.area-f._area)<10);if(j){const k=[];s.left<j.left-2&&k.push("left"),s.right>j.right+2&&k.push("right"),s.top<j.top-2&&k.push("top"),s.bottom>j.bottom+2&&k.push("bottom"),k.length>0&&(m.container={label:f.label,edges:k},g=!0)}}return g?m:null}function u2(s,r){if(s.width/r>.85)return"full-width";const u=s.x+s.width/2,f=r/2,d=u-f,m=r*.08;return Math.abs(d)<m?"center":d<0?"left":"right"}function sy(s){switch(s){case"full-width":return"full-width";case"center":return"centered";case"left":return"left-aligned";case"right":return"right-aligned"}}function ay(s,r={}){const c=[];s.above&&c.push(`Below \`${s.above.label}\`${s.above.gap>0?` (${s.above.gap}px gap)`:""}`),s.below&&c.push(`Above \`${s.below.label}\`${s.below.gap>0?` (${s.below.gap}px gap)`:""}`),r.includeLeftRight&&(s.left&&c.push(`Right of \`${s.left.label}\`${s.left.gap>0?` (${s.left.gap}px gap)`:""}`),s.right&&c.push(`Left of \`${s.right.label}\`${s.right.gap>0?` (${s.right.gap}px gap)`:""}`));const u=sy(s.alignment);return s.containedIn?c.push(`${u.charAt(0).toUpperCase()+u.slice(1)} in \`${s.containedIn.label}\``):c.push(`${u.charAt(0).toUpperCase()+u.slice(1)} in page`),r.includePixelRef&&r.pixelRef&&c.push(`Pixel ref: \`${r.pixelRef}\``),s.outOfBounds&&(s.outOfBounds.viewport&&c.push(`**Outside viewport** (${s.outOfBounds.viewport.join(", ")} edge${s.outOfBounds.viewport.length>1?"s":""})`),s.outOfBounds.container&&c.push(`**Outside \`${s.outOfBounds.container.label}\`** (${s.outOfBounds.container.edges.join(", ")} edge${s.outOfBounds.container.edges.length>1?"s":""})`)),c}function d2(s,r,c){var d,m;const u=[];s.above&&u.push(`below \`${s.above.label}\``),s.below&&u.push(`above \`${s.below.label}\``),s.left&&u.push(`right of \`${s.left.label}\``),s.right&&u.push(`left of \`${s.right.label}\``),s.containedIn&&u.push(`inside \`${s.containedIn.label}\``),u.push(sy(s.alignment)),(d=s.outOfBounds)!=null&&d.viewport&&u.push(`**outside viewport** (${s.outOfBounds.viewport.join(", ")})`),(m=s.outOfBounds)!=null&&m.container&&u.push(`**outside \`${s.outOfBounds.container.label}\`** (${s.outOfBounds.container.edges.join(", ")})`);const f=c?`, ${Math.round(c.width)}×${Math.round(c.height)}px`:"";return`at (${Math.round(r.x)}, ${Math.round(r.y)})${f}: ${u.join(", ")}`}var W0=15;function K0(s){if(s.length<2)return[];const r=[],c=new Set;for(let u=0;u<s.length;u++){if(c.has(u))continue;const f=[u];for(let d=u+1;d<s.length;d++)c.has(d)||Math.abs(s[u].rect.y-s[d].rect.y)<W0&&f.push(d);if(f.length>=2){const d=f.map(p=>s[p]);d.sort((p,j)=>p.rect.x-j.rect.x);const m=[];for(let p=0;p<d.length-1;p++)m.push(Math.round(d[p+1].rect.x-(d[p].rect.x+d[p].rect.width)));const g=Math.round(d.reduce((p,j)=>p+j.rect.y,0)/d.length);r.push({labels:d.map(p=>p.label),type:"row",sharedEdge:g,gaps:m,avgGap:m.length?Math.round(m.reduce((p,j)=>p+j,0)/m.length):0}),f.forEach(p=>c.add(p))}}for(let u=0;u<s.length;u++){if(c.has(u))continue;const f=[u];for(let d=u+1;d<s.length;d++)c.has(d)||Math.abs(s[u].rect.x-s[d].rect.x)<W0&&f.push(d);if(f.length>=2){const d=f.map(p=>s[p]);d.sort((p,j)=>p.rect.y-j.rect.y);const m=[];for(let p=0;p<d.length-1;p++)m.push(Math.round(d[p+1].rect.y-(d[p].rect.y+d[p].rect.height)));const g=Math.round(d.reduce((p,j)=>p+j.rect.x,0)/d.length);r.push({labels:d.map(p=>p.label),type:"column",sharedEdge:g,gaps:m,avgGap:m.length?Math.round(m.reduce((p,j)=>p+j,0)/m.length):0}),f.forEach(p=>c.add(p))}}return r}function f2(s){if(s.length<2)return[];const r=K0(s.map(m=>({label:m.label,rect:m.originalRect}))),c=K0(s.map(m=>({label:m.label,rect:m.currentRect}))),u=[],f=new Set;for(const m of r){const g=new Set(m.labels);let p=null,j=0;for(const k of c){const b=k.labels.filter(v=>g.has(v)).length;b>=2&&b>j&&(p=k,j=b)}if(p){const k=p.labels.filter(v=>g.has(v)),b=k.join(", ");if(p.type!==m.type){const v=m.type==="row"?"y":"x",z=p.type==="row"?"y":"x";u.push(`**${b}**: ${m.type} (${v}≈${m.sharedEdge}, ${m.avgGap}px gaps) → ${p.type} (${z}≈${p.sharedEdge}, ${p.avgGap}px gaps)`)}else if(Math.abs(m.sharedEdge-p.sharedEdge)>20||Math.abs(m.avgGap-p.avgGap)>5){const v=m.type==="row"?"y":"x",z=Math.abs(m.sharedEdge-p.sharedEdge)>20?` ${v}: ${m.sharedEdge} → ${p.sharedEdge}`:"",E=Math.abs(m.avgGap-p.avgGap)>5?` gaps: ${m.avgGap}px → ${p.avgGap}px`:"";u.push(`**${b}**: ${m.type} shifted —${z}${E}`)}k.forEach(v=>f.add(v))}else{const k=m.labels.join(", "),b=m.type==="row"?"y":"x";u.push(`**${k}**: ${m.type} (${b}≈${m.sharedEdge}) dissolved`),m.labels.forEach(v=>f.add(v))}}for(const m of c){if(m.labels.every(j=>f.has(j))||m.labels.filter(j=>!f.has(j)).length<2)continue;if(!r.some(j=>j.labels.filter(b=>m.labels.includes(b)).length>=2)){const j=m.type==="row"?"y":"x";u.push(`**${m.labels.join(", ")}**: new ${m.type} (${j}≈${m.sharedEdge}, ${m.avgGap}px gaps)`),m.labels.forEach(k=>f.add(k))}}const d=s.filter(m=>!f.has(m.label));if(d.length>=2){const m={};for(const g of d){const p=Math.round(g.currentRect.x/5)*5;(m[p]??(m[p]=[])).push(g.label)}for(const[g,p]of Object.entries(m))p.length>=2&&u.push(`**${p.join(", ")}**: shared left edge at x≈${g}`)}return u}function ry(s){if(typeof document>"u")return{viewport:s,contentArea:null};const r=[],c=new Set,u=g=>{c.has(g)||g instanceof HTMLElement&&(g.hasAttribute("data-feedback-toolbar")||z_.has(g.tagName.toLowerCase())||(c.add(g),r.push(g)))},f=document.querySelector("main");f&&u(f);const d=document.querySelector("[role='main']");d&&u(d);for(const g of Array.from(document.body.children))if(u(g),g.children){for(const p of Array.from(g.children))if(u(p),p.children)for(const j of Array.from(p.children))u(j)}let m=null;for(const g of r){const p=g.getBoundingClientRect();if(p.height<50)continue;const j=getComputedStyle(g);if(j.maxWidth&&j.maxWidth!=="none"&&j.maxWidth!=="0px"){(!m||p.width<m.rect.width)&&(m={el:g,rect:p});continue}!m&&p.width<s.width-20&&p.width>100&&(m={el:g,rect:p})}if(m){const{el:g,rect:p}=m;return{viewport:s,contentArea:{width:Math.round(p.width),left:Math.round(p.left),right:Math.round(p.right),centerX:Math.round(p.left+p.width/2),selector:Ba(g)}}}return{viewport:s,contentArea:null}}function _2(s){if(typeof document>"u")return null;const r=document.querySelector(s);if(!(r!=null&&r.parentElement))return null;const c=getComputedStyle(r.parentElement),u={parentDisplay:c.display,parentSelector:Ba(r.parentElement)};return c.display.includes("flex")&&(u.flexDirection=c.flexDirection),c.display.includes("grid")&&c.gridTemplateColumns!=="none"&&(u.gridCols=c.gridTemplateColumns),c.gap&&c.gap!=="normal"&&c.gap!=="0px"&&(u.gap=c.gap),u}function iy(s,r){const c=r.contentArea,u=c?c.width:r.viewport.width,f=c?c.left:0,d=c?c.centerX:Math.round(r.viewport.width/2),m=Math.round(s.x-f),g=Math.round(f+u-(s.x+s.width)),p=(s.width/u*100).toFixed(1),j=s.x+s.width/2,k=Math.abs(j-d)<20,b=s.width/u>.95,v=[];return b?v.push("`width: 100%` of container"):v.push(`left \`${m}px\` in container, right \`${g}px\`, width \`${p}%\` (\`${Math.round(s.width)}px\`)`),k&&!b&&v.push("centered — `margin-inline: auto`"),v.join(" — ")}function cy(s){const{viewport:r,contentArea:c}=s;let u=`### Reference Frame
`;if(u+=`- Viewport: \`${r.width}×${r.height}px\`
`,c){const f=c;u+=`- Content area: \`${f.width}px\` wide, left edge at \`x=${f.left}\`, right at \`x=${f.right}\` (\`${f.selector}\`)
`,u+=`- Pixel → CSS translation:
`,u+=`  - **Horizontal position in container**: \`element.x - ${f.left}\` → use as \`margin-left\` or \`left\`
`,u+=`  - **Width as % of container**: \`element.width / ${f.width} × 100\` → use as \`width: X%\`
`,u+="  - **Vertical gap between elements**: `nextElement.y - (prevElement.y + prevElement.height)` → use as `margin-top` or `gap`\n",u+=`  - **Centered**: if \`|element.centerX - ${f.centerX}| < 20px\` → use \`margin-inline: auto\`
`}else u+=`- No distinct content container — elements positioned relative to full viewport
`,u+=`- Pixel → CSS translation:
`,u+=`  - **Width as % of viewport**: \`element.width / ${r.width} × 100\` → use as \`width: X%\`
`,u+=`  - **Centered**: if \`|(element.x + element.width/2) - ${Math.round(r.width/2)}| < 20px\` → use \`margin-inline: auto\`
`;return u+=`
`,u}function m2(s){const r=_2(s);if(!r)return null;let c=`\`${r.parentDisplay}\``;return r.flexDirection&&(c+=`, flex-direction: \`${r.flexDirection}\``),r.gridCols&&(c+=`, grid-template-columns: \`${r.gridCols}\``),r.gap&&(c+=`, gap: \`${r.gap}\``),`Parent: ${c} (\`${r.parentSelector}\`)`}function Z0(s,r,c,u="standard"){var B,F,Me,te;if(s.length===0)return"";const f=[...s].sort((C,ie)=>Math.abs(C.y-ie.y)<20?C.x-ie.x:C.y-ie.y);let d="";if(c!=null&&c.blankCanvas?(d+=`## Wireframe: New Page

`,c.wireframePurpose&&(d+=`> **Purpose:** ${c.wireframePurpose}
>
`),d+=`> ${s.length} component${s.length!==1?"s":""} placed — this is a standalone wireframe, not related to the current page.
>
> This wireframe is a rough sketch for exploring ideas.

`):d+=`## Design Layout

> ${s.length} component${s.length!==1?"s":""} placed

`,u==="compact")return d+=`### Components
`,f.forEach((C,ie)=>{var W;const de=((W=mo[C.type])==null?void 0:W.label)||C.type;d+=`${ie+1}. **${de}** — \`${Math.round(C.width)}×${Math.round(C.height)}px\` at \`(${Math.round(C.x)}, ${Math.round(C.y)})\`
`,C.text&&(d+=`   - Note: "${C.text}"
`)}),d;const m=ry(r);d+=cy(m),d+=`### Components
`,f.forEach((C,ie)=>{var V;const de=((V=mo[C.type])==null?void 0:V.label)||C.type,W={x:C.x,y:C.y,width:C.width,height:C.height};d+=`${ie+1}. **${de}** — \`${Math.round(C.width)}×${Math.round(C.height)}px\` at \`(${Math.round(C.x)}, ${Math.round(C.y)})\`
`,C.text&&(d+=`   - Note: "${C.text}"
`);const Ee=A_(W),Ce=ay(Ee,{includeLeftRight:u==="detailed"||u==="forensic"});for(const Ie of Ce)d+=`   - ${Ie}
`;const pe=iy(W,m);pe&&(d+=`   - CSS: ${pe}
`)}),d+=`
### Layout Analysis
`;const g=[];for(const C of f){const ie=g.find(de=>Math.abs(de.y-C.y)<30);ie?ie.items.push(C):g.push({y:C.y,items:[C]})}if(g.sort((C,ie)=>C.y-ie.y),g.forEach((C,ie)=>{C.items.sort((W,Ee)=>W.x-Ee.x);const de=C.items.map(W=>{var Ee;return((Ee=mo[W.type])==null?void 0:Ee.label)||W.type});if(C.items.length===1){const Ee=C.items[0].width>r.width*.8;d+=`- Row ${ie+1} (y≈${Math.round(C.y)}): ${de[0]}${Ee?" — full width":""}
`}else d+=`- Row ${ie+1} (y≈${Math.round(C.y)}): ${de.join(" | ")} — ${C.items.length} items side by side
`}),u==="detailed"||u==="forensic"){d+=`
### Spacing & Gaps
`;for(let C=0;C<f.length-1;C++){const ie=f[C],de=f[C+1],W=((B=mo[ie.type])==null?void 0:B.label)||ie.type,Ee=((F=mo[de.type])==null?void 0:F.label)||de.type,ne=Math.round(de.y-(ie.y+ie.height)),Ce=Math.round(de.x-(ie.x+ie.width));Math.abs(ie.y-de.y)<30?d+=`- ${W} → ${Ee}: \`${Ce}px\` horizontal gap
`:d+=`- ${W} → ${Ee}: \`${ne}px\` vertical gap
`}if(u==="forensic"&&f.length>2){d+=`
### All Pairwise Gaps
`;for(let C=0;C<f.length;C++)for(let ie=C+1;ie<f.length;ie++){const de=f[C],W=f[ie],Ee=((Me=mo[de.type])==null?void 0:Me.label)||de.type,ne=((te=mo[W.type])==null?void 0:te.label)||W.type,Ce=Math.round(W.y-(de.y+de.height)),pe=Math.round(W.x-(de.x+de.width));d+=`- ${Ee} ↔ ${ne}: h=\`${pe}px\` v=\`${Ce}px\`
`}}u==="forensic"&&(d+=`
### Z-Order (placement order)
`,s.forEach((C,ie)=>{var W;const de=((W=mo[C.type])==null?void 0:W.label)||C.type;d+=`${ie}. ${de} at \`(${Math.round(C.x)}, ${Math.round(C.y)})\`
`}))}d+=`
### Suggested Implementation
`;const p=f.some(C=>C.type==="navigation"),j=f.some(C=>C.type==="hero"),k=f.some(C=>C.type==="sidebar"),b=f.some(C=>C.type==="footer"),v=f.filter(C=>C.type==="card"),z=f.filter(C=>C.type==="form"),E=f.filter(C=>C.type==="table"),$=f.filter(C=>C.type==="modal");if(p&&(d+=`- Top navigation bar with logo + nav links + CTA
`),j&&(d+=`- Hero section with heading, subtext, and call-to-action
`),k&&(d+=`- Sidebar layout — use CSS Grid with sidebar + main content area
`),v.length>1?d+=`- ${v.length}-column card grid — use CSS Grid or Flexbox
`:v.length===1&&(d+=`- Card component with image + content area
`),z.length>0&&(d+=`- ${z.length} form${z.length>1?"s":""} — add proper labels, validation, and submit handling
`),E.length>0&&(d+=`- Data table — consider sortable columns and pagination
`),$.length>0&&(d+=`- Modal dialog — add overlay backdrop and focus trapping
`),b&&(d+=`- Multi-column footer with links
`),u==="detailed"||u==="forensic"){if(d+=`
### CSS Suggestions
`,k){const C=f.find(ie=>ie.type==="sidebar");d+=`- \`display: grid; grid-template-columns: ${Math.round(C.width)}px 1fr;\`
`}if(v.length>1){const C=Math.round(v[0].width);d+=`- \`display: grid; grid-template-columns: repeat(${v.length}, ${C}px); gap: 16px;\`
`}p&&(d+="- Navigation: `position: sticky; top: 0; z-index: 50;`\n")}return d}function J0(s,r="standard",c){const{sections:u}=s,f=[];for(const k of u){const b=k.originalRect,v=k.currentRect,z=Math.abs(b.x-v.x)>1||Math.abs(b.y-v.y)>1,E=Math.abs(b.width-v.width)>1||Math.abs(b.height-v.height)>1,$=!!k.note;if(!z&&!E&&!$){r==="forensic"&&f.push({section:k,posMoved:!1,sizeChanged:!1});continue}f.push({section:k,posMoved:z,sizeChanged:E})}if(f.length===0||r!=="forensic"&&f.every(k=>!k.posMoved&&!k.sizeChanged&&!k.section.note))return"";let d=`## Suggested Layout Changes

`;const m=c?c.width:typeof window<"u"?window.innerWidth:0,g=c?c.height:typeof window<"u"?window.innerHeight:0,p=ry({width:m,height:g});r!=="compact"&&(d+=cy(p)),r==="forensic"&&(d+=`> Detected at: \`${new Date(s.detectedAt).toISOString()}\`
`,d+=`> Total sections: ${u.length}

`);const j=k=>u.map(b=>({label:b.label,selector:b.selector,rect:k==="original"?b.originalRect:b.currentRect}));d+=`**Changes:**
`;for(const{section:k,posMoved:b,sizeChanged:v}of f){const z=k.originalRect,E=k.currentRect;if(!b&&!v){k.note?(d+=`- **${k.label}** — note only
`,d+=`  - Note: "${k.note}"
`):d+=`- ${k.label} — unchanged at (${Math.round(E.x)}, ${Math.round(E.y)}) ${Math.round(E.width)}×${Math.round(E.height)}px
`;continue}if(r==="compact"){b&&v?d+=`- Suggested: move **${k.label}** to (${Math.round(E.x)}, ${Math.round(E.y)}) ${Math.round(E.width)}×${Math.round(E.height)}px
`:b?d+=`- Suggested: move **${k.label}** to (${Math.round(E.x)}, ${Math.round(E.y)})
`:d+=`- Suggested: resize **${k.label}** to ${Math.round(E.width)}×${Math.round(E.height)}px
`,k.note&&(d+=`  - Note: "${k.note}"
`);continue}if(b&&v?d+=`- Suggested: move and resize **${k.label}**
`:b?d+=`- Suggested: move **${k.label}**
`:d+=`- Suggested: resize **${k.label}** from ${Math.round(z.width)}×${Math.round(z.height)}px to ${Math.round(E.width)}×${Math.round(E.height)}px
`,k.note&&(d+=`  - Note: "${k.note}"
`),b){const B=A_(z,j("original")),F=A_(E,j("current")),Me=v?{width:z.width,height:z.height}:void 0;d+=`  - Currently ${d2(B,{x:z.x,y:z.y},Me)}
`;const te=v?{width:E.width,height:E.height}:void 0,C=`at (${Math.round(E.x)}, ${Math.round(E.y)})`,ie=te?`, ${Math.round(te.width)}×${Math.round(te.height)}px`:"",W=ay(F,{includeLeftRight:r==="detailed"||r==="forensic"});if(W.length>0){d+=`  - Suggested position ${C}${ie}: ${W[0]}
`;for(let ne=1;ne<W.length;ne++)d+=`    ${W[ne]}
`}else d+=`  - Suggested position ${C}${ie}
`;const Ee=iy(E,p);Ee&&(d+=`  - CSS: ${Ee}
`)}const $=m2(k.selector);if($&&(d+=`  - ${$}
`),d+=`  - Selector: \`${k.selector}\`
`,r==="detailed"||r==="forensic"){const B=k.className?`${k.tagName}.${k.className.split(" ")[0]}`:k.tagName;B!==k.selector&&(d+=`  - Element: \`${B}\`
`),k.role&&(d+=`  - Role: \`${k.role}\`
`),r==="forensic"&&k.textSnippet&&(d+=`  - Text: "${k.textSnippet}"
`)}r==="forensic"&&(d+=`  - Original rect: \`{ x: ${Math.round(z.x)}, y: ${Math.round(z.y)}, w: ${Math.round(z.width)}, h: ${Math.round(z.height)} }\`
`,d+=`  - Current rect: \`{ x: ${Math.round(E.x)}, y: ${Math.round(E.y)}, w: ${Math.round(E.width)}, h: ${Math.round(E.height)} }\`
`)}if(r!=="compact"){const k=f.filter(v=>v.posMoved).map(v=>({label:v.section.label,originalRect:v.section.originalRect,currentRect:v.section.currentRect})),b=f2(k);if(b.length>0){d+=`
### Layout Summary
`;for(const v of b)d+=`- ${v}
`}}if(r!=="compact"&&u.length>1){d+=`
### All Sections (current positions)
`;const k=[...u].sort((b,v)=>Math.abs(b.currentRect.y-v.currentRect.y)<20?b.currentRect.x-v.currentRect.x:b.currentRect.y-v.currentRect.y);for(const b of k){const v=b.currentRect,z=Math.abs(v.x-b.originalRect.x)>1||Math.abs(v.y-b.originalRect.y)>1||Math.abs(v.width-b.originalRect.width)>1||Math.abs(v.height-b.originalRect.height)>1;d+=`- ${b.label}: \`${Math.round(v.width)}×${Math.round(v.height)}px\` at \`(${Math.round(v.x)}, ${Math.round(v.y)})\`${z?" ← suggested":""}
`}}return d}var O_="feedback-annotations-",uy=7;function Z_(s){return`${O_}${s}`}function Zs(s){if(typeof window>"u")return[];try{const r=localStorage.getItem(Z_(s));if(!r)return[];const c=JSON.parse(r),u=Date.now()-uy*24*60*60*1e3;return c.filter(f=>!f.timestamp||f.timestamp>u)}catch{return[]}}function dy(s,r){if(!(typeof window>"u"))try{localStorage.setItem(Z_(s),JSON.stringify(r))}catch{}}function h2(){const s=new Map;if(typeof window>"u")return s;try{const r=Date.now()-uy*24*60*60*1e3;for(let c=0;c<localStorage.length;c++){const u=localStorage.key(c);if(u!=null&&u.startsWith(O_)){const f=u.slice(O_.length),d=localStorage.getItem(u);if(d){const g=JSON.parse(d).filter(p=>!p.timestamp||p.timestamp>r);g.length>0&&s.set(f,g)}}}}catch{}return s}function x_(s,r,c){const u=r.map(f=>({...f,_syncedTo:c}));dy(s,u)}var J_="agentation-design-";function g2(s){if(typeof window>"u")return[];try{const r=localStorage.getItem(`${J_}${s}`);return r?JSON.parse(r):[]}catch{return[]}}function p2(s,r){if(!(typeof window>"u"))try{localStorage.setItem(`${J_}${s}`,JSON.stringify(r))}catch{}}function y2(s){if(!(typeof window>"u"))try{localStorage.removeItem(`${J_}${s}`)}catch{}}var P_="agentation-rearrange-";function x2(s){if(typeof window>"u")return null;try{const r=localStorage.getItem(`${P_}${s}`);return r?JSON.parse(r):null}catch{return null}}function b2(s,r){if(!(typeof window>"u"))try{localStorage.setItem(`${P_}${s}`,JSON.stringify(r))}catch{}}function v2(s){if(!(typeof window>"u"))try{localStorage.removeItem(`${P_}${s}`)}catch{}}var em="agentation-wireframe-";function w2(s){if(typeof window>"u")return null;try{const r=localStorage.getItem(`${em}${s}`);return r?JSON.parse(r):null}catch{return null}}function P0(s,r){if(!(typeof window>"u"))try{localStorage.setItem(`${em}${s}`,JSON.stringify(r))}catch{}}function Qu(s){if(!(typeof window>"u"))try{localStorage.removeItem(`${em}${s}`)}catch{}}var fy="agentation-session-";function tm(s){return`${fy}${s}`}function S2(s){if(typeof window>"u")return null;try{return localStorage.getItem(tm(s))}catch{return null}}function b_(s,r){if(!(typeof window>"u"))try{localStorage.setItem(tm(s),r)}catch{}}function j2(s){if(!(typeof window>"u"))try{localStorage.removeItem(tm(s))}catch{}}var _y=`${fy}toolbar-hidden`;function k2(){if(typeof window>"u")return!1;try{return sessionStorage.getItem(_y)==="1"}catch{return!1}}function C2(s){if(!(typeof window>"u"))try{s&&sessionStorage.setItem(_y,"1")}catch{}}async function v_(s,r){const c=await fetch(`${s}/sessions`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({url:r})});if(!c.ok)throw new Error(`Failed to create session: ${c.status}`);return c.json()}async function ep(s,r){const c=await fetch(`${s}/sessions/${r}`);if(!c.ok)throw new Error(`Failed to get session: ${c.status}`);return c.json()}async function tp(s,r,c){!c.elementPath&&(c.element==="body"||c.element==="html")&&(c={...c,elementPath:c.element});const u=await fetch(`${s}/sessions/${r}/annotations`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(c)});if(!u.ok)throw new Error(`Failed to sync annotation: ${u.status}`);return u.json()}async function w_(s,r,c){const u=await fetch(`${s}/annotations/${r}`,{method:"PATCH",headers:{"Content-Type":"application/json"},body:JSON.stringify(c)});if(!u.ok)throw new Error(`Failed to update annotation: ${u.status}`);return u.json()}async function Gu(s,r){const c=await fetch(`${s}/annotations/${r}`,{method:"DELETE"});if(!c.ok)throw new Error(`Failed to delete annotation: ${c.status}`)}var It={FunctionComponent:0,ClassComponent:1,IndeterminateComponent:2,HostRoot:3,HostPortal:4,HostComponent:5,HostText:6,Fragment:7,Mode:8,ContextConsumer:9,ContextProvider:10,ForwardRef:11,Profiler:12,SuspenseComponent:13,MemoComponent:14,SimpleMemoComponent:15,LazyComponent:16,IncompleteClassComponent:17,DehydratedFragment:18,SuspenseListComponent:19,ScopeComponent:21,OffscreenComponent:22,LegacyHiddenComponent:23,CacheComponent:24,TracingMarkerComponent:25,HostHoistable:26,HostSingleton:27,IncompleteFunctionComponent:28,Throw:29,ViewTransitionComponent:30,ActivityComponent:31},np=new Set(["Component","PureComponent","Fragment","Suspense","Profiler","StrictMode","Routes","Route","Outlet","Root","ErrorBoundaryHandler","HotReload","Hot"]),lp=[/Boundary$/,/BoundaryHandler$/,/Provider$/,/Consumer$/,/^(Inner|Outer)/,/Router$/,/^Client(Page|Segment|Root)/,/^Segment(ViewNode|Node)$/,/^LayoutSegment/,/^Server(Root|Component|Render)/,/^RSC/,/Context$/,/^Hot(Reload)?$/,/^(Dev|React)(Overlay|Tools|Root)/,/Overlay$/,/Handler$/,/^With[A-Z]/,/Wrapper$/,/^Root$/],N2=[/Page$/,/View$/,/Screen$/,/Section$/,/Card$/,/List$/,/Item$/,/Form$/,/Modal$/,/Dialog$/,/Button$/,/Nav$/,/Header$/,/Footer$/,/Layout$/,/Panel$/,/Tab$/,/Menu$/];function M2(s){const r=(s==null?void 0:s.mode)??"filtered";let c=np;if(s!=null&&s.skipExact){const u=s.skipExact instanceof Set?s.skipExact:new Set(s.skipExact);c=new Set([...np,...u])}return{maxComponents:(s==null?void 0:s.maxComponents)??6,maxDepth:(s==null?void 0:s.maxDepth)??30,mode:r,skipExact:c,skipPatterns:s!=null&&s.skipPatterns?[...lp,...s.skipPatterns]:lp,userPatterns:(s==null?void 0:s.userPatterns)??N2,filter:s==null?void 0:s.filter}}function E2(s){return s.replace(/([a-z])([A-Z])/g,"$1-$2").replace(/([A-Z])([A-Z][a-z])/g,"$1-$2").toLowerCase()}function T2(s,r=10){const c=new Set;let u=s,f=0;for(;u&&f<r;)u.className&&typeof u.className=="string"&&u.className.split(/\s+/).forEach(d=>{if(d.length>1){const m=d.replace(/[_][a-zA-Z0-9]{5,}.*$/,"").toLowerCase();m.length>1&&c.add(m)}}),u=u.parentElement,f++;return c}function R2(s,r){const c=E2(s);for(const u of r){if(u===c)return!0;const f=c.split("-").filter(m=>m.length>2),d=u.split("-").filter(m=>m.length>2);for(const m of f)for(const g of d)if(m===g||m.includes(g)||g.includes(m))return!0}return!1}function z2(s,r,c,u){if(c.filter)return c.filter(s,r);switch(c.mode){case"all":return!0;case"filtered":return!(c.skipExact.has(s)||c.skipPatterns.some(f=>f.test(s)));case"smart":return c.skipExact.has(s)||c.skipPatterns.some(f=>f.test(s))?!1:!!(u&&R2(s,u)||c.userPatterns.some(f=>f.test(s)));default:return!0}}var Hr=null,A2=new WeakMap;function S_(s){return Object.keys(s).some(r=>r.startsWith("__reactFiber$")||r.startsWith("__reactInternalInstance$")||r.startsWith("__reactProps$"))}function O2(){if(Hr!==null)return Hr;if(typeof document>"u")return!1;if(document.body&&S_(document.body))return Hr=!0,!0;const s=["#root","#app","#__next","[data-reactroot]"];for(const r of s){const c=document.querySelector(r);if(c&&S_(c))return Hr=!0,!0}if(document.body){for(const r of document.body.children)if(S_(r))return Hr=!0,!0}return Hr=!1,!1}var oc={map:A2};function D2(s){return Object.keys(s).find(c=>c.startsWith("__reactFiber$")||c.startsWith("__reactInternalInstance$"))||null}function L2(s){const r=D2(s);return r?s[r]:null}function Da(s){return s?s.displayName?s.displayName:s.name?s.name:null:null}function B2(s){var f;const{tag:r,type:c,elementType:u}=s;if(r===It.HostComponent||r===It.HostText||r===It.HostHoistable||r===It.HostSingleton||r===It.Fragment||r===It.Mode||r===It.Profiler||r===It.DehydratedFragment||r===It.HostRoot||r===It.HostPortal||r===It.ScopeComponent||r===It.OffscreenComponent||r===It.LegacyHiddenComponent||r===It.CacheComponent||r===It.TracingMarkerComponent||r===It.Throw||r===It.ViewTransitionComponent||r===It.ActivityComponent)return null;if(r===It.ForwardRef){const d=u;if(d!=null&&d.render){const m=Da(d.render);if(m)return m}return d!=null&&d.displayName?d.displayName:Da(c)}if(r===It.MemoComponent||r===It.SimpleMemoComponent){const d=u;if(d!=null&&d.type){const m=Da(d.type);if(m)return m}return d!=null&&d.displayName?d.displayName:Da(c)}if(r===It.ContextProvider){const d=c;return(f=d==null?void 0:d._context)!=null&&f.displayName?`${d._context.displayName}.Provider`:null}if(r===It.ContextConsumer){const d=c;return d!=null&&d.displayName?`${d.displayName}.Consumer`:null}if(r===It.LazyComponent){const d=u;return(d==null?void 0:d._status)===1&&d._result?Da(d._result):null}return r===It.SuspenseComponent||r===It.SuspenseListComponent?null:r===It.IncompleteClassComponent||r===It.IncompleteFunctionComponent||r===It.FunctionComponent||r===It.ClassComponent||r===It.IndeterminateComponent?Da(c):null}function $2(s){return s.length<=2||s.length<=3&&s===s.toLowerCase()}function U2(s,r){const c=M2(r),u=c.mode==="all";if(u){const p=oc.map.get(s);if(p!==void 0)return p}if(!O2()){const p={path:null,components:[]};return u&&oc.map.set(s,p),p}const f=c.mode==="smart"?T2(s):void 0,d=[];try{let p=L2(s),j=0;for(;p&&j<c.maxDepth&&d.length<c.maxComponents;){const k=B2(p);k&&!$2(k)&&z2(k,j,c,f)&&d.push(k),p=p.return,j++}}catch{const p={path:null,components:[]};return u&&oc.map.set(s,p),p}if(d.length===0){const p={path:null,components:[]};return u&&oc.map.set(s,p),p}const g={path:d.slice().reverse().map(p=>`<${p}>`).join(" "),components:d};return u&&oc.map.set(s,g),g}var sc={FunctionComponent:0,IndeterminateComponent:2,ForwardRef:11,MemoComponent:14,SimpleMemoComponent:15};function H2(s){if(!s||typeof s!="object")return null;const r=Object.keys(s),c=r.find(d=>d.startsWith("__reactFiber$"));if(c)return s[c]||null;const u=r.find(d=>d.startsWith("__reactInternalInstance$"));if(u)return s[u]||null;const f=r.find(d=>{if(!d.startsWith("__react"))return!1;const m=s[d];return m&&typeof m=="object"&&"_debugSource"in m});return f&&s[f]||null}function fc(s){if(!s.type||typeof s.type=="string")return null;if(typeof s.type=="object"||typeof s.type=="function"){const r=s.type;if(r.displayName)return r.displayName;if(r.name)return r.name}return null}function Y2(s,r=50){var f;let c=s,u=0;for(;c&&u<r;){if(c._debugSource)return{source:c._debugSource,componentName:fc(c)};if((f=c._debugOwner)!=null&&f._debugSource)return{source:c._debugOwner._debugSource,componentName:fc(c._debugOwner)};c=c.return,u++}return null}function q2(s){let r=s,c=0;const u=50;for(;r&&c<u;){const f=r,d=["_debugSource","__source","_source","debugSource"];for(const m of d){const g=f[m];if(g&&typeof g=="object"&&"fileName"in g)return{source:g,componentName:fc(r)}}if(r.memoizedProps){const m=r.memoizedProps;if(m.__source&&typeof m.__source=="object"){const g=m.__source;if(g.fileName&&g.lineNumber)return{source:{fileName:g.fileName,lineNumber:g.lineNumber,columnNumber:g.columnNumber},componentName:fc(r)}}}r=r.return,c++}return null}var Vu=new Map;function I2(s){var f;const r=s.tag,c=s.type,u=s.elementType;if(typeof c=="string"||c==null||typeof c=="function"&&((f=c.prototype)!=null&&f.isReactComponent))return null;if((r===sc.FunctionComponent||r===sc.IndeterminateComponent)&&typeof c=="function")return c;if(r===sc.ForwardRef&&u){const d=u.render;if(typeof d=="function")return d}if((r===sc.MemoComponent||r===sc.SimpleMemoComponent)&&u){const d=u.type;if(typeof d=="function")return d}return typeof c=="function"?c:null}function X2(){const s=jb,r=s.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE;if(r&&"H"in r)return{get:()=>r.H,set:u=>{r.H=u}};const c=s.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED;if(c){const u=c.ReactCurrentDispatcher;if(u&&"current"in u)return{get:()=>u.current,set:f=>{u.current=f}}}return null}function Q2(s,r){const c=s.split(`
`),u=[/source-location/,/\/dist\/index\./,/node_modules\//,/react-dom/,/react\.development/,/react\.production/,/chunk-[A-Z0-9]+/i,/\/_next\/static\/chunks\//,/\/\.vite\/deps\//,/\/_astro\//,/\/assets\/[^\s/]+[-.][\w-]{8,}\.m?js(?:[?:]|$)/,/react-stack-bottom-frame/,/react-reconciler/,/scheduler/,/<anonymous>/],f=/^\s*at\s+(?:.*?\s+\()?(.+?):(\d+):(\d+)\)?$/,d=/^[^@]*@(.+?):(\d+):(\d+)$/;for(const m of c){const g=m.trim();if(!g||u.some(j=>j.test(g)))continue;if(r){const j=r.replace(/^bound /,"").replace(/[.*+?^${}()|[\]\\]/g,"\\$&");if(!new RegExp(`(?:at (?:Object\\.)?|^)${j}(?: \\(|@| \\[)`).test(g))continue}const p=f.exec(g)||d.exec(g);if(p)return{fileName:p[1],line:parseInt(p[2],10),column:parseInt(p[3],10)}}return null}function G2(s){let r=s;return r=r.replace(/[?#].*$/,""),r=r.replace(/^turbopack:\/\/\/\[project\]\//,""),r=r.replace(/^webpack-internal:\/\/\/\.\//,""),r=r.replace(/^webpack-internal:\/\/\//,""),r=r.replace(/^webpack:\/\/\/\.\//,""),r=r.replace(/^webpack:\/\/\//,""),r=r.replace(/^turbopack:\/\/\//,""),r=r.replace(/^https?:\/\/[^/]+\//,""),r=r.replace(/^file:\/\/\//,"/"),r=r.replace(/^\([^)]+\)\/\.\//,""),r=r.replace(/^\.\//,""),r}function V2(s){const r=I2(s);if(!r)return null;if(Vu.has(r))return Vu.get(r);const c=X2();if(!c)return Vu.set(r,null),null;const u=c.get();let f=null;try{const d=new Proxy({},{get(){throw new Error("probe")}});c.set(d);try{r({})}catch(m){if(m instanceof Error&&m.message==="probe"&&m.stack){const g=Q2(m.stack,r.name);g&&(f={fileName:G2(g.fileName),lineNumber:g.line,columnNumber:g.column,componentName:fc(s)||void 0})}}}finally{c.set(u)}return Vu.set(r,f),f}function F2(s,r=15){let c=s,u=0;for(;c&&u<r;){const f=V2(c);if(f)return f;c=c.return,u++}return null}function D_(s){const r=H2(s);if(!r)return{found:!1,reason:"no-fiber",isReactApp:!1,isProduction:!1};let c=Y2(r);if(c||(c=q2(r)),c!=null&&c.source)return{found:!0,source:{fileName:c.source.fileName,lineNumber:c.source.lineNumber,columnNumber:c.source.columnNumber,componentName:c.componentName||void 0},isReactApp:!0,isProduction:!1};const u=F2(r);return u?{found:!0,source:u,isReactApp:!0,isProduction:!1}:{found:!1,reason:"no-debug-source",isReactApp:!0,isProduction:!1}}function W2(s,r="path"){const{fileName:c,lineNumber:u,columnNumber:f}=s;let d=`${c}:${u}`;return f!==void 0&&(d+=`:${f}`),r==="vscode"?`vscode://file${c.startsWith("/")?"":"/"}${d}`:d}function K2(s,r=10){let c=s,u=0;for(;c&&u<r;){const f=D_(c);if(f.found)return f;c=c.parentElement,u++}return D_(s)}var ac=[{value:"compact",label:"Compact"},{value:"standard",label:"Standard"},{value:"detailed",label:"Detailed"},{value:"forensic",label:"Forensic"}];function td(s,r){let c=`## Page Feedback: ${s}
`;const u=r==null?void 0:r.replace(/[\r\n\t]+/g," ").trim();return u&&(c+=`**App:** ${u.replace(/[\\`*_\[\]<>]/g,"\\$&")}
`),c}function op(s,r,c="standard",u={}){if(s.length===0)return"";const f=typeof window<"u"?`${window.innerWidth}×${window.innerHeight}`:"unknown";let d=td(r,u.appName);return c==="forensic"?(d+=`
**Environment:**
`,d+=`- Viewport: ${f}
`,typeof window<"u"&&(d+=`- URL: ${window.location.href}
`,d+=`- User Agent: ${navigator.userAgent}
`,d+=`- Timestamp: ${new Date().toISOString()}
`,d+=`- Device Pixel Ratio: ${window.devicePixelRatio}
`),d+=`
---
`):c!=="compact"&&(d+=`**Viewport:** ${f}
`),d+=`
`,s.forEach((m,g)=>{c==="compact"?(d+=`${g+1}. **${m.element}**${m.sourceFile?` (${m.sourceFile})`:""}: ${m.comment}`,m.selectedText&&(d+=` (re: "${m.selectedText.slice(0,30)}${m.selectedText.length>30?"...":""}")`),d+=`
`):c==="forensic"?(d+=`### ${g+1}. ${m.element}
`,m.isMultiSelect&&m.fullPath&&(d+=`*Forensic data shown for first element of selection*
`),m.fullPath&&(d+=`**Full DOM Path:** ${m.fullPath}
`),m.cssClasses&&(d+=`**CSS Classes:** ${m.cssClasses}
`),m.boundingBox&&(d+=`**Position:** x:${Math.round(m.boundingBox.x)}, y:${Math.round(m.boundingBox.y)} (${Math.round(m.boundingBox.width)}×${Math.round(m.boundingBox.height)}px)
`),d+=`**Annotation at:** ${m.x.toFixed(1)}% from left, ${Math.round(m.y)}px from top
`,m.selectedText&&(d+=`**Selected text:** "${m.selectedText}"
`),m.nearbyText&&!m.selectedText&&(d+=`**Context:** ${m.nearbyText.slice(0,100)}
`),m.computedStyles&&(d+=`**Computed Styles:** ${m.computedStyles}
`),m.accessibility&&(d+=`**Accessibility:** ${m.accessibility}
`),m.nearbyElements&&(d+=`**Nearby Elements:** ${m.nearbyElements}
`),m.sourceFile&&(d+=`**Source:** ${m.sourceFile}
`),m.reactComponents&&(d+=`**React:** ${m.reactComponents}
`),d+=`**Feedback:** ${m.comment}

`):(d+=`### ${g+1}. ${m.element}
`,d+=`**Location:** ${m.elementPath}
`,m.sourceFile&&(d+=`**Source:** ${m.sourceFile}
`),m.reactComponents&&(d+=`**React:** ${m.reactComponents}
`),c==="detailed"&&(m.cssClasses&&(d+=`**Classes:** ${m.cssClasses}
`),m.boundingBox&&(d+=`**Position:** ${Math.round(m.boundingBox.x)}px, ${Math.round(m.boundingBox.y)}px (${Math.round(m.boundingBox.width)}×${Math.round(m.boundingBox.height)}px)
`)),m.selectedText&&(d+=`**Selected text:** "${m.selectedText}"
`),c==="detailed"&&m.nearbyText&&!m.selectedText&&(d+=`**Context:** ${m.nearbyText.slice(0,100)}
`),d+=`**Feedback:** ${m.comment}

`)}),d.trim()}function sp(s,r,c="markdown"){return c==="markdown"?r:[...new Set(s.map(u=>{var f;return c==="source"?u.sourceFile:c==="classes"?u.cssClasses:(f=u.attributes)==null?void 0:f[c.attribute]}).filter(u=>typeof u=="string"&&u.length>0))].join(`
`)}async function Z2(s){var r;if(typeof window>"u")return!1;try{if((r=navigator.clipboard)!=null&&r.writeText)return await navigator.clipboard.writeText(s),!0}catch{}return J2(s)}function J2(s){var g;const r=document.createElement("textarea");let c=document.activeElement;for(;(g=c==null?void 0:c.shadowRoot)!=null&&g.activeElement;)c=c.shadowRoot.activeElement;const u=c instanceof HTMLInputElement||c instanceof HTMLTextAreaElement?c:null,f=u&&u.selectionStart!==null?{start:u.selectionStart,end:u.selectionEnd,direction:u.selectionDirection}:null,d=document.getSelection(),m=d?Array.from({length:d.rangeCount},(p,j)=>d.getRangeAt(j).cloneRange()):[];try{return r.value=s,r.setAttribute("readonly",""),r.style.cssText="position:fixed;left:-9999px;top:0;opacity:0;pointer-events:none;",document.body.appendChild(r),r.focus({preventScroll:!0}),r.select(),r.setSelectionRange(0,s.length),document.execCommand("copy")}catch{return!1}finally{if(r.remove(),c instanceof HTMLElement&&c.isConnected&&(c.focus({preventScroll:!0}),u&&f&&u.setSelectionRange(f.start,f.end,f.direction)),d){d.removeAllRanges();for(const p of m)d.addRange(p)}}}function ap(s){if(!s)return s;try{const r=new URL(s,"http://agentation.invalid");return r.pathname+r.search+r.hash}catch{return s}}function rp(s,r,c=[]){const u=new Map,f=new Map(c.map(g=>[g.id,g]));let d=!1;async function m(g,p){if(d||p.running||p.timer)return;const j=p.desired,k=r.get(g);if(!j&&!k){u.delete(g),r.delete(g);return}const b=j?JSON.stringify(j):void 0;if(!(j&&k&&p.synced===b)){p.running=!0;try{if(!j)await s.remove(k),r.delete(g),p.synced=void 0;else if(k)await s.update(k,j),p.synced=b;else{r.set(g,"");const v=await s.create(j);r.set(g,v.id),p.synced=b}p.retries=0,p.running=!1,!d&&u.get(g)===p&&m(g,p)}catch(v){if(p.running=!1,!d&&u.get(g)===p&&(console.warn("[Agentation] Failed to sync layout feedback:",v),r.get(g)&&p.retries<3)){const z=500*2**p.retries++;p.timer=Ht(()=>{p.timer=void 0,m(g,p)},z)}}}}return{replace(g){const p=new Map(g.map(j=>[j.id,j]));for(const[j,k]of p){let b=u.get(j);if(!b){b={running:!1,retries:0},u.set(j,b);const v=[...f.values()].find(z=>z.kind===k.kind&&ap(z.url)===ap(k.url)&&(k.kind==="placement"?z.timestamp===k.timestamp&&z.element===k.element:z.element===k.element));v&&(r.set(j,v.id),f.delete(v.id))}JSON.stringify(b.desired)!==JSON.stringify(k)&&(b.retries=0,b.timer&&clearTimeout(b.timer),b.timer=void 0),b.desired=k}for(const[j,k]of u)p.has(j)||(k.desired=void 0),m(j,k)},forget(g){const p=u.get(g);p!=null&&p.timer&&clearTimeout(p.timer),u.delete(g),r.delete(g)},dispose(){d=!0;for(const g of u.values())g.timer&&clearTimeout(g.timer)}}}function P2(s,r,c,u){let f=!1,d,m,g=1e3,p;const j=z=>{!f&&((z==null?void 0:z.status)==="resolved"||(z==null?void 0:z.status)==="dismissed")&&u(z)},k=async()=>{if(f||p||!c())return;const z=new AbortController;p=z;const E=Ht(()=>z.abort(),5e3);try{const $=await fetch(`${s}/sessions/${r}`,{signal:z.signal});if(!$.ok)return;const B=await $.json();!f&&!z.signal.aborted&&Array.isArray(B.annotations)&&B.annotations.forEach(j)}catch{}finally{clearTimeout(E),p===z&&(p=void 0)}},b=()=>{if(f)return;const z=new EventSource(`${s}/sessions/${r}/events`),E=()=>{g=1e3,k()},$=F=>{try{j(JSON.parse(F.data).payload)}catch{}},B=()=>{z.readyState!==EventSource.CLOSED||f||m!==void 0||(d==null||d(),m=Ht(()=>{m=void 0,b()},g),g=Math.min(g*2,1e4))};z.addEventListener("open",E),z.addEventListener("annotation.updated",$),z.addEventListener("error",B),d=()=>{z.removeEventListener("open",E),z.removeEventListener("annotation.updated",$),z.removeEventListener("error",B),z.close()}};b();const v=Qp(()=>{k()},1e4);return()=>{f=!0,d==null||d(),m!==void 0&&clearTimeout(m),clearInterval(v),p==null||p.abort()}}var ew=`.styles-module__surface___7qnpJ {
  padding: 0;
  width: var(--preview-width, 200px);
  max-width: calc(100vw - 24px);
  overflow: auto;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  border-radius: 12px;
  z-index: inherit;
  will-change: auto;
  transform: translateX(-50%);
  transition: left 200ms cubic-bezier(0.2, 0.8, 0.2, 1), top 200ms cubic-bezier(0.2, 0.8, 0.2, 1), transform 200ms cubic-bezier(0.2, 0.8, 0.2, 1), width 200ms cubic-bezier(0.2, 0.8, 0.2, 1), opacity 100ms ease-out, visibility 0s 200ms;
}
.styles-module__surface___7qnpJ[data-positioning] {
  transition: none;
}
.styles-module__surface___7qnpJ[data-direct-entry] *, .styles-module__surface___7qnpJ[data-direct-entry] *::before, .styles-module__surface___7qnpJ[data-direct-entry] *::after {
  transition: none !important;
}
.styles-module__surface___7qnpJ[data-state=preview], .styles-module__surface___7qnpJ[data-state=edit] {
  opacity: 1;
  visibility: visible;
  transition-delay: 0s;
}
.styles-module__surface___7qnpJ[data-state=edit], .styles-module__surface___7qnpJ[data-annotation-popup][data-state=hidden] {
  width: 280px;
  border-radius: 16px;
}
.styles-module__surface___7qnpJ[data-state=edit] {
  pointer-events: auto;
}
.styles-module__surface___7qnpJ[data-state=preview] {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.08);
}
[data-agentation-theme=light] .styles-module__surface___7qnpJ[data-state=preview] {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(0, 0, 0, 0.06);
}

@media (prefers-reduced-motion: reduce) {
  .styles-module__surface___7qnpJ {
    transition: opacity 100ms ease-out, visibility 0s 100ms;
  }
}`,tw={surface:"styles-module__surface___7qnpJ"},nw=h.forwardRef(function({annotation:r,editing:c,exiting:u,restorePreview:f,editorProps:d,lightMode:m,scrollY:g,onExited:p},j){const k=h.useRef({annotation:r,editorProps:d}),b=r??k.current.annotation,v=r?d:k.current.editorProps,z=h.useRef(null),E=h.useRef(null),$=h.useRef(),B=h.useRef(g),F=c&&!u?"edit":r&&(!c||f)?"preview":"hidden",Me=F==="preview"||!c;return h.useLayoutEffect(()=>{r&&(k.current={annotation:r,editorProps:d})},[r,d]),h.useLayoutEffect(()=>{const te=z.current;if(!te||!b)return;const C=F==="edit"&&($.current!==b.id||te.dataset.state==="hidden"&&getComputedStyle(te).opacity==="0");C&&(te.dataset.directEntry="true");const ie=()=>{const Ee=b.x/100*window.innerWidth,ne=b.isFixed?b.y:b.y-g,Ce=F==="preview"||!c,pe=parseFloat(te.style.getPropertyValue("--preview-width"))||200,V=Math.min(pe,window.innerWidth-24),Ie=Math.max(12,Math.min(window.innerWidth-V-12,Ee-V/2)),he=Ce?V:Math.min(280,window.innerWidth-24),ue=Math.min(Ce?12:20,(window.innerWidth-he)/2),G=ne>window.innerHeight-(Ce?101:290);te.style.left=`${Math.max(ue,Math.min(window.innerWidth-he-ue,Ie))}px`,te.style.right="auto",te.style.top=`${Math.max(12,Math.min(window.innerHeight-12,ne+(G?-21:21)))}px`,te.style.bottom="auto",te.style.transform=G?"translateY(-100%)":"translateY(0)",te.style.maxHeight=`${Math.max(100,G?ne-33:window.innerHeight-ne-33)}px`},de=$.current!==b.id||B.current!==g||te.dataset.state==="hidden";de&&(te.dataset.positioning="true"),ie(),(de||C)&&te.getBoundingClientRect(),delete te.dataset.positioning,delete te.dataset.directEntry,te.dataset.state=F,te.inert=F!=="edit",$.current=b.id,B.current=g;const W=()=>{te.dataset.positioning="true",ie(),te.getBoundingClientRect(),delete te.dataset.positioning};return window.addEventListener("resize",W),()=>window.removeEventListener("resize",W)},[b==null?void 0:b.id,b==null?void 0:b.x,b==null?void 0:b.y,b==null?void 0:b.isFixed,F,c,g,b==null?void 0:b.comment]),h.useLayoutEffect(()=>{var te;c&&!u&&((te=E.current)==null||te.focus())},[c,u,b==null?void 0:b.id]),F_(z,u,p),h.useImperativeHandle(j,()=>({shake(){var te,C,ie;(C=(te=z.current)==null?void 0:te.animate)==null||C.call(te,[{translate:"0px"},{translate:"-3px"},{translate:"3px"},{translate:"-2px"},{translate:"2px"},{translate:"0px"}],{duration:250}),(ie=E.current)==null||ie.focus()}}),[]),!b||!v?null:o.jsx("div",{ref:z,className:`${xt.popup} ${tw.surface} ${m?xt.light:""}`,"data-feedback-toolbar":!0,"data-annotation-card":!0,"data-annotation-popup":c?"":void 0,"data-state":"hidden","aria-hidden":F==="hidden",onClick:te=>te.stopPropagation(),onKeyDownCapture:te=>{te.key!=="Escape"||te.nativeEvent.isComposing||!c||(te.preventDefault(),te.stopPropagation(),v.onCancel())},children:o.jsx(Pp,{ref:E,...v,variant:"card",preview:Me,resetOnPreview:!c,disabled:!c||u},b.id)})}),lw=`@keyframes styles-module__markerIn___x4G8D {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.3);
  }
  100% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
}
@keyframes styles-module__markerOut___6VhQN {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.3);
  }
}
@keyframes styles-module__renumberRoll___akV9B {
  0% {
    transform: translateX(-40%);
    opacity: 0;
  }
  100% {
    transform: translateX(0);
    opacity: 1;
  }
}
.styles-module__marker___9CKF7 {
  padding: 0;
  border: 0;
  box-sizing: border-box;
  font-family: inherit;
  line-height: 1;
  text-align: center;
  appearance: none;
  position: absolute;
  width: 22px;
  height: 22px;
  background: var(--agentation-color-blue);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6875rem;
  font-weight: 600;
  transform: translate(-50%, -50%) scale(1);
  opacity: 1;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2), inset 0 0 0 1px rgba(0, 0, 0, 0.04);
  -webkit-user-select: none;
  user-select: none;
  will-change: transform, opacity;
  contain: layout style;
  z-index: 1;
}
.styles-module__marker___9CKF7 > * {
  pointer-events: none;
}
.styles-module__marker___9CKF7:focus-visible {
  outline: 2px solid var(--agentation-color-accent);
  outline-offset: 3px;
}
.styles-module__marker___9CKF7:hover, .styles-module__marker___9CKF7:focus-visible, .styles-module__marker___9CKF7.styles-module__previewVisible___imMag {
  z-index: 2;
}
.styles-module__marker___9CKF7:not(.styles-module__enter___8kI3q):not(.styles-module__exit___KBdR3):not(.styles-module__clearing___8rM7K):not(.styles-module__confirm___BtMvq) {
  transition: background-color 0.15s ease, transform 0.1s ease, z-index 0s 0.1s;
}
.styles-module__marker___9CKF7:not(.styles-module__enter___8kI3q):not(.styles-module__exit___KBdR3):not(.styles-module__clearing___8rM7K):not(.styles-module__confirm___BtMvq):hover, .styles-module__marker___9CKF7:not(.styles-module__enter___8kI3q):not(.styles-module__exit___KBdR3):not(.styles-module__clearing___8rM7K):not(.styles-module__confirm___BtMvq):focus-visible, .styles-module__marker___9CKF7:not(.styles-module__enter___8kI3q):not(.styles-module__exit___KBdR3):not(.styles-module__clearing___8rM7K):not(.styles-module__confirm___BtMvq).styles-module__previewVisible___imMag {
  transition-delay: 0s;
}
.styles-module__marker___9CKF7.styles-module__enter___8kI3q {
  animation: styles-module__markerIn___x4G8D 0.25s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.styles-module__marker___9CKF7.styles-module__confirm___BtMvq {
  animation: styles-module__markerConfirm___RT4Sk 220ms ease-out both;
}
.styles-module__marker___9CKF7.styles-module__exit___KBdR3 {
  animation: styles-module__markerOut___6VhQN 0.2s ease-out both;
  pointer-events: none;
}
.styles-module__marker___9CKF7.styles-module__clearing___8rM7K {
  animation: styles-module__markerOut___6VhQN 0.15s ease-out both;
  pointer-events: none;
}
.styles-module__marker___9CKF7:not(.styles-module__enter___8kI3q):not(.styles-module__exit___KBdR3):not(.styles-module__clearing___8rM7K):not(.styles-module__confirm___BtMvq):hover {
  transform: translate(-50%, -50%) scale(1.1);
}
.styles-module__marker___9CKF7.styles-module__pending___BiY-U {
  background-color: var(--agentation-color-blue);
  cursor: default;
}
.styles-module__marker___9CKF7.styles-module__pending___BiY-U.styles-module__exit___KBdR3 {
  animation-duration: 150ms;
}
.styles-module__marker___9CKF7.styles-module__multiSelect___CPfTC {
  background-color: var(--agentation-color-green);
  width: 26px;
  height: 26px;
  border-radius: 6px;
  font-size: 0.75rem;
}
.styles-module__marker___9CKF7.styles-module__multiSelect___CPfTC.styles-module__pending___BiY-U {
  background-color: var(--agentation-color-green);
}
.styles-module__marker___9CKF7.styles-module__hovered___-mg2N {
  background-color: var(--agentation-color-red);
}

.styles-module__renumber___16lvD {
  display: block;
  animation: styles-module__renumberRoll___akV9B 0.2s ease-out;
}

@keyframes styles-module__markerConfirm___RT4Sk {
  0% {
    transform: translate(-50%, -50%) scale(1);
  }
  25% {
    transform: translate(-50%, -50%) scale(0.94);
  }
  65% {
    transform: translate(-50%, -50%) scale(1.06);
  }
  100% {
    transform: translate(-50%, -50%) scale(1);
  }
}
.styles-module__number___1JFu9 {
  display: block;
}

.styles-module__numberGlyph___qchdk {
  display: block;
  opacity: 1;
  transform: translateY(0);
  filter: blur(0);
  transition: opacity 140ms ease-out, transform 180ms cubic-bezier(0.22, 1, 0.36, 1), filter 140ms ease-out;
}

.styles-module__actionGlyph___AFRt0 {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  opacity: 0;
  transform: translateY(2px) scale(0.8) rotate(-12deg);
  filter: blur(1px);
  transition: opacity 120ms ease-out, transform 160ms cubic-bezier(0.22, 1, 0.36, 1), filter 120ms ease-out;
}

.styles-module__actionVisible___Kb--l .styles-module__numberGlyph___qchdk {
  opacity: 0;
  transform: translateY(-2px) scale(0.8);
  filter: blur(1px);
}
.styles-module__actionVisible___Kb--l .styles-module__actionGlyph___AFRt0 {
  opacity: 1;
  transform: translateY(0) scale(1) rotate(0);
  filter: blur(0);
}

.styles-module__plus___xslMP {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  opacity: 0;
  transform: rotate(-90deg) scale(0.6);
  filter: blur(1px);
  transition: opacity 100ms ease-out, transform 140ms ease-out, filter 100ms ease-out;
  pointer-events: none;
}

.styles-module__pending___BiY-U .styles-module__numberGlyph___qchdk {
  opacity: 0;
  transform: translateY(5px);
  filter: blur(2px);
}
.styles-module__pending___BiY-U .styles-module__plus___xslMP {
  opacity: 1;
  transform: rotate(0) scale(1);
  filter: blur(0);
}

@media (prefers-reduced-motion: reduce) {
  .styles-module__marker___9CKF7.styles-module__enter___8kI3q, .styles-module__marker___9CKF7.styles-module__exit___KBdR3, .styles-module__marker___9CKF7.styles-module__clearing___8rM7K, .styles-module__marker___9CKF7.styles-module__confirm___BtMvq {
    animation-duration: 1ms;
    animation-delay: 0ms !important;
  }
  .styles-module__numberGlyph___qchdk, .styles-module__actionGlyph___AFRt0, .styles-module__plus___xslMP {
    transition: opacity 100ms ease-out;
    transform: none;
    filter: none;
  }
  .styles-module__pending___BiY-U .styles-module__numberGlyph___qchdk, .styles-module__pending___BiY-U .styles-module__plus___xslMP,
  .styles-module__actionVisible___Kb--l .styles-module__numberGlyph___qchdk, .styles-module__actionVisible___Kb--l .styles-module__actionGlyph___AFRt0 {
    transform: none;
    filter: none;
  }
}`,ll={marker:"styles-module__marker___9CKF7",previewVisible:"styles-module__previewVisible___imMag",enter:"styles-module__enter___8kI3q",exit:"styles-module__exit___KBdR3",clearing:"styles-module__clearing___8rM7K",confirm:"styles-module__confirm___BtMvq",pending:"styles-module__pending___BiY-U",multiSelect:"styles-module__multiSelect___CPfTC",hovered:"styles-module__hovered___-mg2N",renumber:"styles-module__renumber___16lvD",number:"styles-module__number___1JFu9",numberGlyph:"styles-module__numberGlyph___qchdk",actionGlyph:"styles-module__actionGlyph___AFRt0",actionVisible:"styles-module__actionVisible___Kb--l",plus:"styles-module__plus___xslMP"},ip=h.memo(function({annotation:r,pending:c=!1,globalIndex:u,layerIndex:f,layerSize:d,isExiting:m,isClearing:g,isAnimated:p,isNew:j,isHovered:k,isRemoving:b,onRemoveComplete:v,isEditingAny:z,renumberFrom:E,markerClickBehavior:$,onHoverEnter:B,onEnterComplete:F,onHoverLeave:Me,onClick:te,onContextMenu:C}){const[ie,de]=h.useState(p),W=h.useRef(c),[Ee,ne]=h.useState(!1),Ce=W.current&&!c&&ie&&!Ee;h.useLayoutEffect(()=>{m&&de(!1)},[m]);const pe=h.useRef(null),V=h.useRef({action:!1,delete:!1}),Ie=k&&!z,he=Ie&&$==="delete";h.useLayoutEffect(()=>{b||(V.current={action:Ie,delete:he})},[b,Ie,he]);const ue=b?V.current.action:Ie,G=b?V.current.delete:he;F_(pe,b,()=>v(r.id));const we=r.isMultiSelect,fe=we?"var(--agentation-color-green)":"var(--agentation-color-accent)",Oe=g?ll.clearing:m||b?ll.exit:Ce?ll.confirm:!p&&!ie?ll.enter:"",Se=g?`${Math.min(f*20,120)}ms`:b||c||Ce?"0ms":m?`${(d-1-f)*20}ms`:`${j?0:f*20}ms`;return o.jsxs("button",{ref:pe,type:"button","aria-label":c?"Pending annotation":`${$==="delete"?"Delete":"Edit"} annotation ${u+1}: ${r.element}`,disabled:c||m||b||g,tabIndex:c||z?-1:0,className:`${ll.marker} ${c?ll.pending:""} ${we?ll.multiSelect:""} ${Oe} ${!c&&ue?ll.actionVisible:""} ${G?ll.hovered:""} ${k&&!z&&!b?ll.previewVisible:""}`,"data-annotation-marker":c?void 0:"","data-annotation-pending":c?"":void 0,style:{left:`${r.x}%`,top:r.y,backgroundColor:G?void 0:fe,animationDelay:Se},onAnimationEnd:Qe=>{Qe.target===Qe.currentTarget&&(Oe===ll.enter||Oe===ll.confirm)&&(de(!0),c||ne(!0),c||F(r.id))},onMouseOver:()=>{c||B(r)},onMouseOut:Qe=>{const Ut=Qe.relatedTarget;(!(Ut instanceof Node)||!Qe.currentTarget.contains(Ut))&&Me(r.id)},onFocus:Qe=>{!c&&Qe.currentTarget.matches(":focus-visible")&&B(r)},onBlur:()=>Me(r.id),onClick:Qe=>{Qe.stopPropagation(),!c&&!m&&!b&&te(r,Qe.currentTarget)},onContextMenu:C?Qe=>{$==="delete"&&(Qe.preventDefault(),Qe.stopPropagation(),!c&&!m&&!b&&C(r,Qe.currentTarget))}:void 0,children:[o.jsx("span",{className:`${ll.number} ${E!==null&&u>=E?ll.renumber:""}`,"aria-hidden":"true",children:o.jsx("span",{className:ll.numberGlyph,children:u+1})},u),o.jsx("span",{className:ll.actionGlyph,"aria-hidden":"true",children:$==="delete"?o.jsx(wv,{size:we?18:16}):o.jsx(kv,{size:16})}),W.current&&o.jsx("span",{className:ll.plus,"aria-hidden":"true",children:o.jsx(mv,{size:12})})]})}),ow=`.styles-module__switchContainer___Ka-AB {
  display: flex;
  align-items: center;
  position: relative;
  padding: 2px;
  width: 24px;
  height: 16px;
  border-radius: 8px;
  background-color: #cdcdcd;
  transition: background-color 0.15s, opacity 0.15s;
}
[data-agentation-theme=dark] .styles-module__switchContainer___Ka-AB {
  background-color: #484848;
}
.styles-module__switchContainer___Ka-AB:has(.styles-module__switchInput___kYDSD:checked) {
  background-color: var(--agentation-color-blue);
}
.styles-module__switchContainer___Ka-AB:has(.styles-module__switchInput___kYDSD:disabled) {
  opacity: 0.3;
}

.styles-module__switchInput___kYDSD {
  position: absolute;
  z-index: 1;
  inset: 0;
  border-radius: inherit;
  opacity: 0;
  cursor: pointer;
}
.styles-module__switchInput___kYDSD:disabled {
  cursor: not-allowed;
}

.styles-module__switchThumb___4sCPH {
  border-radius: 50%;
  width: 12px;
  height: 12px;
  background-color: #fff;
  transition: transform 0.15s;
}
.styles-module__switchContainer___Ka-AB[data-checked] .styles-module__switchThumb___4sCPH {
  transform: translateX(8px);
}`,j_={switchContainer:"styles-module__switchContainer___Ka-AB",switchInput:"styles-module__switchInput___kYDSD",switchThumb:"styles-module__switchThumb___4sCPH"},k_=({className:s="",checked:r,onChange:c,...u})=>o.jsxs("div",{className:`${j_.switchContainer} ${s}`,"data-checked":r?"":void 0,children:[o.jsx("input",{className:j_.switchInput,checked:r,onChange:c,type:"checkbox",...u}),o.jsx("div",{className:j_.switchThumb})]}),sw=`.styles-module__checkboxContainer___joqZk {
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  border: 1px solid rgba(26, 26, 26, 0.2);
  border-radius: 4px;
  width: 14px;
  height: 14px;
  background-color: #fff;
  transition: background-color 0.2s ease;
}
[data-agentation-theme=dark] .styles-module__checkboxContainer___joqZk {
  border-color: rgba(255, 255, 255, 0.2);
  background-color: #252525;
}
.styles-module__checkboxContainer___joqZk:has(.styles-module__checkboxInput___ECzzO:checked) {
  background-color: #1a1a1a;
}
[data-agentation-theme=dark] .styles-module__checkboxContainer___joqZk:has(.styles-module__checkboxInput___ECzzO:checked) {
  background-color: #fff;
}

.styles-module__checkboxInput___ECzzO {
  position: absolute;
  z-index: 1;
  inset: -1px;
  border-radius: inherit;
  opacity: 0;
  cursor: pointer;
}

.styles-module__checkboxCheck___fUXpr {
  color: #fafafa;
}
[data-agentation-theme=dark] .styles-module__checkboxCheck___fUXpr {
  color: #1a1a1a;
}

.styles-module__checkboxCheckPath___cDyh8 {
  stroke-dasharray: 9.29px;
  stroke-dashoffset: 9.29px;
  color: #fafafa;
  transition: stroke-dashoffset 0.1s ease;
}
[data-agentation-theme=dark] .styles-module__checkboxCheckPath___cDyh8 {
  color: #1a1a1a;
}
.styles-module__checkboxContainer___joqZk[data-checked] .styles-module__checkboxCheckPath___cDyh8 {
  transition-duration: 0.2s;
  stroke-dashoffset: 0;
}`,Fu={checkboxContainer:"styles-module__checkboxContainer___joqZk",checkboxInput:"styles-module__checkboxInput___ECzzO",checkboxCheck:"styles-module__checkboxCheck___fUXpr",checkboxCheckPath:"styles-module__checkboxCheckPath___cDyh8"},aw=({className:s="",checked:r,onChange:c,...u})=>o.jsxs("div",{className:`${Fu.checkboxContainer} ${s}`,"data-checked":r?"":void 0,children:[o.jsx("input",{className:Fu.checkboxInput,type:"checkbox",checked:r,onChange:c,...u}),o.jsx("svg",{className:Fu.checkboxCheck,width:"14",height:"14",viewBox:"0 0 14 14",fill:"none",children:o.jsx("path",{className:Fu.checkboxCheckPath,d:"M3.94 7L6.13 9.19L10.5 4.81",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})})]}),rw=`.styles-module__container___w8eAF {
  display: flex;
  align-items: center;
  height: 24px;
}

.styles-module__label___J5mxE {
  padding-inline: 8px 2px;
  line-height: 20px;
  font-size: 13px;
  letter-spacing: -0.15px;
  color: rgba(26, 26, 26, 0.5);
  -webkit-user-select: none;
  user-select: none;
  cursor: pointer;
}
[data-agentation-theme=dark] .styles-module__label___J5mxE {
  color: rgba(255, 255, 255, 0.5);
}`,cp={container:"styles-module__container___w8eAF",label:"styles-module__label___J5mxE"},up=({className:s="",label:r,tooltip:c,checked:u,onChange:f,...d})=>{const m=h.useId();return o.jsxs("div",{className:`${cp.container} ${s}`,...d,children:[o.jsx(aw,{id:m,onChange:f,checked:u}),o.jsx("label",{className:cp.label,htmlFor:m,children:r}),c&&o.jsx(La,{content:c})]})},iw=`@keyframes styles-module__cycleTextIn___VBNTi {
  0% {
    opacity: 0;
    transform: translateY(-6px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}
@keyframes styles-module__scaleIn___QpQ8E {
  from {
    opacity: 0;
    transform: scale(0.85);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
@keyframes styles-module__mcpPulse___5Q3Jj {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-green) 50%, transparent);
  }
  70% {
    box-shadow: 0 0 0 6px color-mix(in srgb, var(--agentation-color-green) 0%, transparent);
  }
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-green) 0%, transparent);
  }
}
@keyframes styles-module__mcpPulseError___VHxhx {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-red) 50%, transparent);
  }
  70% {
    box-shadow: 0 0 0 6px color-mix(in srgb, var(--agentation-color-red) 0%, transparent);
  }
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--agentation-color-red) 0%, transparent);
  }
}
@keyframes styles-module__themeIconIn___qUWMV {
  0% {
    opacity: 0;
    transform: scale(0.8) rotate(-30deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}
.styles-module__settingsPanel___qNkn- :where(button, a, input, select, textarea):focus-visible {
  outline: 2px solid var(--agentation-color-accent);
  outline-offset: 2px;
}
.styles-module__settingsPanel___qNkn- {
  position: absolute;
  right: 5px;
  bottom: calc(100% + 0.5rem);
  z-index: 1;
  overflow: hidden;
  background: #1c1c1c;
  border-radius: 16px;
  padding: 12px 0;
  width: 253px;
  max-width: calc(100vw - 20px);
  cursor: default;
  opacity: 1;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(0, 0, 0, 0.04);
  transition: background-color 0.25s ease, box-shadow 0.25s ease;
}
.styles-module__settingsPanel___qNkn-::before, .styles-module__settingsPanel___qNkn-::after {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  width: 16px;
  z-index: 2;
  pointer-events: none;
}
.styles-module__settingsPanel___qNkn-::before {
  left: 0;
  background: linear-gradient(to right, #1c1c1c 0%, transparent 100%);
}
.styles-module__settingsPanel___qNkn-::after {
  right: 0;
  background: linear-gradient(to left, #1c1c1c 0%, transparent 100%);
}
.styles-module__settingsPanel___qNkn- .styles-module__settingsHeader___Fn1DP,
.styles-module__settingsPanel___qNkn- .styles-module__settingsBrand___OoKlM,
.styles-module__settingsPanel___qNkn- .styles-module__settingsVersion___rXmL9,
.styles-module__settingsPanel___qNkn- .styles-module__settingsSection___n5V-4,
.styles-module__settingsPanel___qNkn- .styles-module__settingsLabel___VCVOQ,
.styles-module__settingsPanel___qNkn- .styles-module__cycleButton___XMBx3,
.styles-module__settingsPanel___qNkn- .styles-module__cycleDot___zgSXY,
.styles-module__settingsPanel___qNkn- .styles-module__dropdownButton___mKHe8,
.styles-module__settingsPanel___qNkn- .styles-module__sliderLabel___6K5v1,
.styles-module__settingsPanel___qNkn- .styles-module__slider___v5z-c,
.styles-module__settingsPanel___qNkn- .styles-module__themeToggle___3imlT {
  transition: background-color 0.25s ease, color 0.25s ease, border-color 0.25s ease;
}
.styles-module__settingsPanel___qNkn- {
  opacity: 0;
  transform: translateY(var(--panel-offset-y, 4px)) scale(0.98);
  transform-origin: var(--panel-origin, bottom right);
  filter: blur(2px);
  pointer-events: none;
  visibility: hidden;
  transition: opacity 120ms cubic-bezier(0.25, 0.46, 0.45, 0.94), transform 120ms cubic-bezier(0.25, 0.46, 0.45, 0.94), filter 120ms cubic-bezier(0.25, 0.46, 0.45, 0.94);
}
.styles-module__settingsPanel___qNkn-[data-panel-present=true] {
  visibility: visible;
}
.styles-module__settingsPanel___qNkn-[data-panel-open=true] {
  opacity: 1;
  transform: translateY(0) scale(1);
  filter: blur(0);
  pointer-events: auto;
  transition-duration: 160ms;
}
@media (prefers-reduced-motion: reduce) {
  .styles-module__settingsPanel___qNkn- {
    transition: none;
    transform: none;
    filter: none;
  }
}
.styles-module__settingsPanel___qNkn-.styles-module__below___Vpv-k {
  --panel-offset-y: -4px;
  --panel-origin: top right;
}
[data-agentation-theme=dark] .styles-module__settingsPanel___qNkn- {
  background: #1a1a1a;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.08);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___qNkn- .styles-module__settingsLabel___VCVOQ {
  color: rgba(255, 255, 255, 0.6);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___qNkn- .styles-module__settingsOption___JoyH- {
  color: rgba(255, 255, 255, 0.85);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___qNkn- .styles-module__settingsOption___JoyH-:hover {
  background: rgba(255, 255, 255, 0.1);
}
[data-agentation-theme=dark] .styles-module__settingsPanel___qNkn- .styles-module__settingsOption___JoyH-.styles-module__selected___k1-Vq {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

.styles-module__settingsPanelContainer___5it-H {
  overflow: visible;
  position: relative;
  display: flex;
  padding: 0 16px;
}

.styles-module__settingsPage___BMn-3 {
  min-width: 100%;
  flex-basis: 0;
  flex-shrink: 0;
  transition: transform 0.2s ease, opacity 0.2s ease;
  transition-delay: 0s;
  opacity: 1;
}

.styles-module__settingsPage___BMn-3.styles-module__slideLeft___qUvW4 {
  transform: translateX(-24px);
  opacity: 0;
  pointer-events: none;
}

.styles-module__automationsPage___N7By0 {
  position: absolute;
  top: 0;
  left: 24px;
  width: 100%;
  height: 100%;
  padding: 0 16px 4px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  transition: transform 0.2s ease, opacity 0.2s ease;
  opacity: 0;
  pointer-events: none;
}

.styles-module__automationsPage___N7By0.styles-module__slideIn___uXDSu {
  transform: translateX(-24px);
  opacity: 1;
  pointer-events: auto;
}

.styles-module__settingsHeader___Fn1DP {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 24px;
}

.styles-module__settingsBrand___OoKlM {
  display: flex;
  align-items: center;
  font-size: 0.8125rem;
  font-weight: 500;
  letter-spacing: -0.0094em;
  color: #bbb;
  text-decoration: none;
}

.styles-module__settingsVersion___rXmL9 {
  font-size: 11px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.4);
  margin-left: 6px;
  letter-spacing: -0.0094em;
}

.styles-module__themeToggle___3imlT {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  margin-left: auto;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: rgba(255, 255, 255, 0.4);
  transition: background-color 0.15s ease, color 0.15s ease;
  cursor: pointer;
}
.styles-module__themeToggle___3imlT:hover {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}
[data-agentation-theme=light] .styles-module__themeToggle___3imlT {
  color: rgba(0, 0, 0, 0.4);
}
[data-agentation-theme=light] .styles-module__themeToggle___3imlT:hover {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.7);
}

.styles-module__themeIconWrapper___pyaYa {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  width: 20px;
  height: 20px;
}

.styles-module__themeIcon___w7lAm {
  display: flex;
  align-items: center;
  justify-content: center;
  animation: styles-module__themeIconIn___qUWMV 0.35s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
}

.styles-module__settingsSectionGrow___eZTRw {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.styles-module__settingsRow___y-tDE {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 24px;
}
.styles-module__settingsRow___y-tDE.styles-module__settingsRowMarginTop___uLpGb {
  margin-top: 8px;
}

.styles-module__settingsRowDisabled___ydl3Q .styles-module__settingsLabel___VCVOQ {
  color: rgba(255, 255, 255, 0.2);
}
[data-agentation-theme=light] .styles-module__settingsRowDisabled___ydl3Q .styles-module__settingsLabel___VCVOQ {
  color: rgba(0, 0, 0, 0.2);
}

.styles-module__settingsLabel___VCVOQ {
  display: flex;
  align-items: center;
  column-gap: 2px;
  line-height: 20px;
  font-size: 13px;
  font-weight: 400;
  letter-spacing: -0.15px;
  color: rgba(255, 255, 255, 0.5);
}
[data-agentation-theme=light] .styles-module__settingsLabel___VCVOQ {
  color: rgba(0, 0, 0, 0.5);
}

.styles-module__cycleButton___XMBx3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0;
  border: none;
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #fff;
  cursor: pointer;
  letter-spacing: -0.0094em;
}
[data-agentation-theme=light] .styles-module__cycleButton___XMBx3 {
  color: rgba(0, 0, 0, 0.85);
}
.styles-module__cycleButton___XMBx3:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.styles-module__cycleButtonText___mbbnD {
  display: inline-block;
  animation: styles-module__cycleTextIn___VBNTi 0.2s ease-out;
}

.styles-module__cycleDots___ehp6i {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.styles-module__cycleDot___zgSXY {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  transform: scale(0.667);
  transition: background-color 0.25s ease-out, transform 0.25s ease-out;
}
.styles-module__cycleDot___zgSXY.styles-module__active___dpAhM {
  background: #fff;
  transform: scale(1);
}
[data-agentation-theme=light] .styles-module__cycleDot___zgSXY {
  background: rgba(0, 0, 0, 0.2);
}
[data-agentation-theme=light] .styles-module__cycleDot___zgSXY.styles-module__active___dpAhM {
  background: rgba(0, 0, 0, 0.7);
}

.styles-module__colorOptions___pbxZx {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
  height: 26px;
}

.styles-module__colorOption___Co955 {
  padding: 0;
  position: relative;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  background-color: #fff;
  cursor: pointer;
}
[data-agentation-theme=dark] .styles-module__colorOption___Co955 {
  background-color: #1a1a1a;
}
.styles-module__colorOption___Co955::before, .styles-module__colorOption___Co955::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background-color: var(--swatch);
  transition: opacity 0.2s, transform 0.2s;
}
@supports (color: color(display-p3 0 0 0)) {
  .styles-module__colorOption___Co955::before, .styles-module__colorOption___Co955::after {
    --color: var(--swatch-p3);
  }
}
.styles-module__colorOption___Co955::after {
  z-index: -1;
  transform: scale(1.2);
  opacity: 0;
}
.styles-module__colorOption___Co955.styles-module__selected___k1-Vq::before {
  transform: scale(0.8);
}
.styles-module__colorOption___Co955.styles-module__selected___k1-Vq::after {
  opacity: 1;
}

.styles-module__settingsNavLink___uYIwM {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 24px;
  padding: 0;
  border: none;
  background: transparent;
  font-family: inherit;
  line-height: 20px;
  font-size: 13px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.5);
  transition: color 0.15s ease;
  cursor: pointer;
}
.styles-module__settingsNavLink___uYIwM:hover {
  color: rgba(255, 255, 255, 0.9);
}
.styles-module__settingsNavLink___uYIwM svg {
  color: rgba(255, 255, 255, 0.4);
  transition: color 0.15s ease;
}
.styles-module__settingsNavLink___uYIwM:hover svg {
  color: #fff;
}
[data-agentation-theme=light] .styles-module__settingsNavLink___uYIwM {
  color: rgba(0, 0, 0, 0.5);
}
[data-agentation-theme=light] .styles-module__settingsNavLink___uYIwM:hover {
  color: rgba(0, 0, 0, 0.8);
}
[data-agentation-theme=light] .styles-module__settingsNavLink___uYIwM svg {
  color: rgba(0, 0, 0, 0.25);
}
[data-agentation-theme=light] .styles-module__settingsNavLink___uYIwM:hover svg {
  color: rgba(0, 0, 0, 0.8);
}

.styles-module__settingsNavLinkRight___XBUzC {
  display: flex;
  align-items: center;
  gap: 6px;
}

.styles-module__settingsBackButton___fflll {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 24px;
  border: none;
  background: transparent;
  font-family: inherit;
  line-height: 20px;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: -0.15px;
  color: #fff;
  cursor: pointer;
  transition: transform 0.12s cubic-bezier(0.32, 0.72, 0, 1);
}
.styles-module__settingsBackButton___fflll svg {
  opacity: 0.4;
  flex-shrink: 0;
  transition: opacity 0.15s ease, transform 0.18s cubic-bezier(0.32, 0.72, 0, 1);
}
.styles-module__settingsBackButton___fflll:hover svg {
  opacity: 1;
}
[data-agentation-theme=light] .styles-module__settingsBackButton___fflll {
  color: rgba(0, 0, 0, 0.85);
  border-bottom-color: rgba(0, 0, 0, 0.08);
}

.styles-module__automationHeader___Avra9 {
  display: flex;
  align-items: center;
  gap: 0.125rem;
  font-size: 0.8125rem;
  font-weight: 400;
  color: #fff;
}
[data-agentation-theme=light] .styles-module__automationHeader___Avra9 {
  color: rgba(0, 0, 0, 0.85);
}

.styles-module__automationDescription___vFTmJ {
  font-size: 0.6875rem;
  font-weight: 300;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 2px;
  line-height: 14px;
}
[data-agentation-theme=light] .styles-module__automationDescription___vFTmJ {
  color: rgba(0, 0, 0, 0.5);
}

.styles-module__learnMoreLink___cG7OI {
  color: rgba(255, 255, 255, 0.8);
  text-decoration-line: underline;
  text-decoration-style: dotted;
  text-decoration-color: rgba(255, 255, 255, 0.2);
  text-underline-offset: 2px;
  transition: color 0.15s ease;
}
.styles-module__learnMoreLink___cG7OI:hover {
  color: #fff;
}
[data-agentation-theme=light] .styles-module__learnMoreLink___cG7OI {
  color: rgba(0, 0, 0, 0.6);
  text-decoration-color: rgba(0, 0, 0, 0.2);
}
[data-agentation-theme=light] .styles-module__learnMoreLink___cG7OI:hover {
  color: rgba(0, 0, 0, 0.85);
}

.styles-module__autoSendContainer___VpkXk {
  display: flex;
  align-items: center;
}

.styles-module__autoSendLabel___ngNdC {
  padding-inline-end: 8px;
  font-size: 11px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.4);
  transition: color 0.15s, opacity 0.15s;
  cursor: pointer;
}
.styles-module__autoSendLabel___ngNdC.styles-module__active___dpAhM {
  color: #66b8ff;
  color: color(display-p3 0.4 0.72 1);
}
[data-agentation-theme=light] .styles-module__autoSendLabel___ngNdC {
  color: rgba(0, 0, 0, 0.4);
}
[data-agentation-theme=light] .styles-module__autoSendLabel___ngNdC.styles-module__active___dpAhM {
  color: var(--agentation-color-blue);
}
.styles-module__autoSendLabel___ngNdC.styles-module__disabled___9AZYS {
  opacity: 0.3;
  cursor: not-allowed;
}

.styles-module__mcpStatusDot___8AMxP {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.styles-module__mcpStatusDot___8AMxP.styles-module__connecting___QEO1r {
  background-color: var(--agentation-color-yellow);
  animation: styles-module__mcpPulse___5Q3Jj 1.5s infinite;
}
.styles-module__mcpStatusDot___8AMxP.styles-module__connected___WyFkx {
  background-color: var(--agentation-color-green);
  animation: styles-module__mcpPulse___5Q3Jj 2.5s ease-in-out infinite;
}
.styles-module__mcpStatusDot___8AMxP.styles-module__disconnected___mvmvQ {
  background-color: var(--agentation-color-red);
  animation: styles-module__mcpPulseError___VHxhx 2s infinite;
}

.styles-module__mcpNavIndicator___auBHI {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.styles-module__mcpNavIndicator___auBHI.styles-module__connected___WyFkx {
  background-color: var(--agentation-color-green);
  animation: styles-module__mcpPulse___5Q3Jj 2.5s ease-in-out infinite;
}
.styles-module__mcpNavIndicator___auBHI.styles-module__connecting___QEO1r {
  background-color: var(--agentation-color-yellow);
  animation: styles-module__mcpPulse___5Q3Jj 1.5s ease-in-out infinite;
}

.styles-module__webhookUrlInput___WDDDC {
  display: block;
  width: 100%;
  flex: 1;
  min-height: 60px;
  box-sizing: border-box;
  margin-top: 11px;
  padding: 8px 10px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.03);
  font-family: inherit;
  font-size: 0.75rem;
  font-weight: 400;
  color: #fff;
  outline: none;
  resize: none;
  user-select: text;
  transition: border-color 0.15s ease, background-color 0.15s ease, box-shadow 0.15s ease;
}
.styles-module__webhookUrlInput___WDDDC::placeholder {
  color: rgba(255, 255, 255, 0.3);
}
.styles-module__webhookUrlInput___WDDDC:focus {
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.08);
}
[data-agentation-theme=light] .styles-module__webhookUrlInput___WDDDC {
  border-color: rgba(0, 0, 0, 0.1);
  background: rgba(0, 0, 0, 0.03);
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__webhookUrlInput___WDDDC::placeholder {
  color: rgba(0, 0, 0, 0.3);
}
[data-agentation-theme=light] .styles-module__webhookUrlInput___WDDDC:focus {
  border-color: rgba(0, 0, 0, 0.25);
  background: rgba(0, 0, 0, 0.05);
}

[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08), 0 4px 16px rgba(0, 0, 0, 0.06), 0 0 0 1px rgba(0, 0, 0, 0.04);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn-::before {
  background: linear-gradient(to right, #fff 0%, transparent 100%);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn-::after {
  background: linear-gradient(to left, #fff 0%, transparent 100%);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__settingsHeader___Fn1DP {
  border-bottom-color: rgba(0, 0, 0, 0.08);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__settingsBrand___OoKlM {
  color: #333;
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__settingsVersion___rXmL9 {
  color: rgba(0, 0, 0, 0.4);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__settingsSection___n5V-4 {
  border-top-color: rgba(0, 0, 0, 0.08);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__settingsLabel___VCVOQ {
  color: rgba(0, 0, 0, 0.5);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__cycleButton___XMBx3 {
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__cycleDot___zgSXY {
  background: rgba(0, 0, 0, 0.2);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__cycleDot___zgSXY.styles-module__active___dpAhM {
  background: rgba(0, 0, 0, 0.7);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__dropdownButton___mKHe8 {
  color: rgba(0, 0, 0, 0.85);
}
[data-agentation-theme=light] .styles-module__settingsPanel___qNkn- .styles-module__dropdownButton___mKHe8:hover {
  background: rgba(0, 0, 0, 0.05);
}

.styles-module__checkboxField___ZrSqv:not(:first-child) {
  margin-top: 8px;
}

.styles-module__divider___h6Yux {
  margin-block: 8px;
  width: 100%;
  height: 1px;
  background-color: rgba(26, 26, 26, 0.07);
}
[data-agentation-theme=dark] .styles-module__divider___h6Yux {
  background-color: rgba(255, 255, 255, 0.07);
}`,De={settingsPanel:"styles-module__settingsPanel___qNkn-",settingsHeader:"styles-module__settingsHeader___Fn1DP",settingsBrand:"styles-module__settingsBrand___OoKlM",settingsVersion:"styles-module__settingsVersion___rXmL9",settingsSection:"styles-module__settingsSection___n5V-4",settingsLabel:"styles-module__settingsLabel___VCVOQ",cycleButton:"styles-module__cycleButton___XMBx3",cycleDot:"styles-module__cycleDot___zgSXY",dropdownButton:"styles-module__dropdownButton___mKHe8",sliderLabel:"styles-module__sliderLabel___6K5v1",slider:"styles-module__slider___v5z-c",themeToggle:"styles-module__themeToggle___3imlT",below:"styles-module__below___Vpv-k",settingsOption:"styles-module__settingsOption___JoyH-",selected:"styles-module__selected___k1-Vq",settingsPanelContainer:"styles-module__settingsPanelContainer___5it-H",settingsPage:"styles-module__settingsPage___BMn-3",slideLeft:"styles-module__slideLeft___qUvW4",automationsPage:"styles-module__automationsPage___N7By0",slideIn:"styles-module__slideIn___uXDSu",themeIconWrapper:"styles-module__themeIconWrapper___pyaYa",themeIcon:"styles-module__themeIcon___w7lAm",themeIconIn:"styles-module__themeIconIn___qUWMV",settingsSectionGrow:"styles-module__settingsSectionGrow___eZTRw",settingsRow:"styles-module__settingsRow___y-tDE",settingsRowMarginTop:"styles-module__settingsRowMarginTop___uLpGb",settingsRowDisabled:"styles-module__settingsRowDisabled___ydl3Q",cycleButtonText:"styles-module__cycleButtonText___mbbnD",cycleTextIn:"styles-module__cycleTextIn___VBNTi",cycleDots:"styles-module__cycleDots___ehp6i",active:"styles-module__active___dpAhM",colorOptions:"styles-module__colorOptions___pbxZx",colorOption:"styles-module__colorOption___Co955",settingsNavLink:"styles-module__settingsNavLink___uYIwM",settingsNavLinkRight:"styles-module__settingsNavLinkRight___XBUzC",settingsBackButton:"styles-module__settingsBackButton___fflll",automationHeader:"styles-module__automationHeader___Avra9",automationDescription:"styles-module__automationDescription___vFTmJ",learnMoreLink:"styles-module__learnMoreLink___cG7OI",autoSendContainer:"styles-module__autoSendContainer___VpkXk",autoSendLabel:"styles-module__autoSendLabel___ngNdC",disabled:"styles-module__disabled___9AZYS",mcpStatusDot:"styles-module__mcpStatusDot___8AMxP",connecting:"styles-module__connecting___QEO1r",mcpPulse:"styles-module__mcpPulse___5Q3Jj",connected:"styles-module__connected___WyFkx",disconnected:"styles-module__disconnected___mvmvQ",mcpPulseError:"styles-module__mcpPulseError___VHxhx",mcpNavIndicator:"styles-module__mcpNavIndicator___auBHI",webhookUrlInput:"styles-module__webhookUrlInput___WDDDC",checkboxField:"styles-module__checkboxField___ZrSqv",divider:"styles-module__divider___h6Yux",scaleIn:"styles-module__scaleIn___QpQ8E"},cw=h.memo(function({settings:r,onSettingsChange:c,isDarkMode:u,onToggleTheme:f,isDevMode:d,connectionStatus:m,endpoint:g,onExited:p,isOpen:j,toolbarNearBottom:k,settingsPage:b,onSettingsPageChange:v,onHideToolbar:z}){var te;const{ref:E}=ty(j,{keepMounted:!0,onExited:p}),$=h.useRef(null),B=h.useRef(null),F=h.useRef(!1);h.useLayoutEffect(()=>{var C;!j||!F.current||(F.current=!1,(C=(b==="automations"?B:$).current)==null||C.focus())},[j,b]);const Me=u?"Switch to light mode":"Switch to dark mode";return o.jsx("div",{className:`${De.settingsPanel} ${k?De.below:""}`,style:k?{bottom:"auto",top:"calc(100% + 0.5rem)"}:void 0,"data-agentation-settings-panel":!0,ref:C=>{E.current=C,C==null||C.toggleAttribute("inert",!j)},role:"group","aria-label":"Feedback settings","aria-hidden":!j,children:o.jsxs("div",{className:De.settingsPanelContainer,children:[o.jsxs("div",{className:`${De.settingsPage} ${b==="automations"?De.slideLeft:""}`,ref:C=>{C==null||C.toggleAttribute("inert",b!=="main")},"aria-hidden":b!=="main",children:[o.jsxs("div",{className:De.settingsHeader,children:[o.jsx("a",{className:De.settingsBrand,href:"https://agentation.com",target:"_blank",rel:"noopener noreferrer","aria-label":"Agentation",children:"Agentation"}),o.jsxs("p",{className:De.settingsVersion,children:["v","3.1.2"]}),o.jsx("button",{className:De.themeToggle,onClick:f,title:Me,"aria-label":Me,children:o.jsx("span",{className:De.themeIconWrapper,children:o.jsx("span",{className:De.themeIcon,children:u?o.jsx(Sv,{size:20}):o.jsx(jv,{size:20})},u?"sun":"moon")})})]}),o.jsx("div",{className:De.divider}),o.jsxs("div",{className:De.settingsSection,children:[o.jsxs("div",{className:De.settingsRow,children:[o.jsxs("div",{className:De.settingsLabel,children:["Output Detail",o.jsx(La,{content:"Controls how much detail is included in the copied output"})]}),o.jsxs("button",{className:De.cycleButton,onClick:()=>{const ie=(ac.findIndex(de=>de.value===r.outputDetail)+1)%ac.length;c({outputDetail:ac[ie].value})},children:[o.jsx("span",{className:De.cycleButtonText,children:(te=ac.find(C=>C.value===r.outputDetail))==null?void 0:te.label},r.outputDetail),o.jsx("span",{className:De.cycleDots,children:ac.map(C=>o.jsx("span",{className:`${De.cycleDot} ${r.outputDetail===C.value?De.active:""}`},C.value))})]})]}),o.jsxs("div",{className:`${De.settingsRow} ${De.settingsRowMarginTop} ${d?"":De.settingsRowDisabled}`,children:[o.jsxs("div",{className:De.settingsLabel,children:["React Components",o.jsx(La,{content:d?"Include React component names in annotations":"Disabled — production builds minify component names, making detection unreliable. Use in development mode."})]}),o.jsx(k_,{"aria-label":"React Components",checked:d&&r.reactEnabled,onChange:C=>c({reactEnabled:C.target.checked}),disabled:!d})]}),o.jsxs("div",{className:`${De.settingsRow} ${De.settingsRowMarginTop}`,children:[o.jsxs("div",{className:De.settingsLabel,children:["Hide Until Restart",o.jsx(La,{content:"Hides the toolbar until you open a new tab"})]}),o.jsx(k_,{"aria-label":"Hide Until Restart",checked:!1,onChange:C=>{C.target.checked&&z()}})]})]}),o.jsx("div",{className:De.divider}),o.jsxs("div",{className:De.settingsSection,children:[o.jsx("div",{className:`${De.settingsLabel} ${De.settingsLabelMarker}`,children:"Marker Color"}),o.jsx("div",{className:De.colorOptions,children:dc.map(C=>o.jsx("button",{className:`${De.colorOption} ${r.annotationColorId===C.id?De.selected:""}`,style:{"--swatch":C.srgb,"--swatch-p3":C.p3},onClick:()=>c({annotationColorId:C.id}),title:C.label,"aria-label":C.label,type:"button"},C.id))})]}),o.jsx("div",{className:De.divider}),o.jsxs("div",{className:De.settingsSection,children:[o.jsx(up,{className:"checkbox-field",label:"Clear on copy/send",checked:r.autoClearAfterCopy,onChange:C=>c({autoClearAfterCopy:C.target.checked}),tooltip:"Automatically clear annotations after copying"}),o.jsx(up,{className:De.checkboxField,label:"Block page interactions",checked:r.blockInteractions,onChange:C=>c({blockInteractions:C.target.checked})})]}),o.jsx("div",{className:De.divider}),o.jsxs("button",{className:De.settingsNavLink,ref:$,onClick:C=>{F.current=C.detail===0,C.currentTarget.blur(),v("automations")},children:[o.jsx("span",{children:"Manage MCP & Webhooks"}),o.jsxs("span",{className:De.settingsNavLinkRight,children:[g&&m!=="disconnected"&&o.jsx("span",{className:`${De.mcpNavIndicator} ${De[m]}`}),o.jsx("svg",{width:"16",height:"16",viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg",children:o.jsx("path",{d:"M7.5 12.5L12 8L7.5 3.5",stroke:"currentColor",strokeWidth:"1.5",strokeLinecap:"round",strokeLinejoin:"round"})})]})]})]}),o.jsxs("div",{className:`${De.settingsPage} ${De.automationsPage} ${b==="automations"?De.slideIn:""}`,ref:C=>{C==null||C.toggleAttribute("inert",b!=="automations")},"aria-hidden":b!=="automations",children:[o.jsxs("button",{className:De.settingsBackButton,ref:B,"aria-label":"Back to settings",onClick:C=>{F.current=C.detail===0,C.currentTarget.blur(),v("main")},children:[o.jsx(Cv,{size:16}),o.jsx("span",{children:"Manage MCP & Webhooks"})]}),o.jsx("div",{className:De.divider}),o.jsxs("div",{className:De.settingsSection,children:[o.jsxs("div",{className:De.settingsRow,children:[o.jsxs("span",{className:De.automationHeader,children:["MCP Connection",o.jsx(La,{content:"Connect via Model Context Protocol to let AI agents like Claude Code receive annotations in real-time."})]}),g&&o.jsx("div",{className:`${De.mcpStatusDot} ${De[m]}`,title:m==="connected"?"Connected":m==="connecting"?"Connecting...":"Disconnected"})]}),o.jsxs("p",{className:De.automationDescription,style:{paddingBottom:6},children:["MCP connection allows agents to receive and act on annotations."," ",o.jsx("a",{href:"https://agentation.com/mcp",target:"_blank",rel:"noopener noreferrer",className:De.learnMoreLink,children:"Learn more"})]})]}),o.jsx("div",{className:De.divider}),o.jsxs("div",{className:`${De.settingsSection} ${De.settingsSectionGrow}`,children:[o.jsxs("div",{className:De.settingsRow,children:[o.jsxs("span",{className:De.automationHeader,children:["Webhooks",o.jsx(La,{content:"Send annotation data to any URL endpoint when annotations change. Useful for custom integrations."})]}),o.jsxs("div",{className:De.autoSendContainer,children:[o.jsx("label",{htmlFor:"agentation-auto-send",className:`${De.autoSendLabel} ${r.webhooksEnabled?De.active:""} ${r.webhookUrl?"":De.disabled}`,children:"Auto-Send"}),o.jsx(k_,{id:"agentation-auto-send",checked:r.webhooksEnabled,onChange:C=>c({webhooksEnabled:C.target.checked}),disabled:!r.webhookUrl})]})]}),o.jsx("p",{className:De.automationDescription,children:"The webhook URL will receive live annotation changes and annotation data."}),o.jsx("textarea",{className:De.webhookUrlInput,placeholder:"Webhook URL","aria-label":"Webhook URL",value:r.webhookUrl,onKeyDown:C=>C.stopPropagation(),onChange:C=>c({webhookUrl:C.target.value})})]})]})]})})});function uw({x:s,y:r,elementName:c,reactComponents:u}){const f=h.useRef(null);return h.useLayoutEffect(()=>{const d=f.current;if(!d)return;const m=()=>{const g=d.offsetWidth,p=d.offsetHeight;d.style.left=`${Math.max(8,Math.min(s,window.innerWidth-g-8))}px`;const j=r-p-8;d.style.top=`${Math.max(8,Math.min(j,window.innerHeight-p-8))}px`};return m(),window.addEventListener("resize",m),()=>window.removeEventListener("resize",m)},[s,r,c,u]),o.jsxs("div",{ref:f,className:`${le.hoverTooltip} ${le.enter}`,children:[u&&o.jsx("div",{className:le.hoverReactPath,children:u}),o.jsx("div",{className:le.hoverElementName,children:c})]})}var dw=`@charset "UTF-8";
/* Reset box-model and set borders */
/* ============================================ */
*,
::before,
::after {
  border-width: 0;
  border-style: solid;
  box-sizing: border-box;
}

/* Document */
/* ============================================ */
/**
 * 1. Correct line height in all browsers.
 * 2. Prevent adjustments of font size after orientation changes in iOS.
 * 3. Remove gray overlay on links for iOS.
 * 4. Render kerning consistently in all browsers.
 * 5. Correct font smoothing for macOS.
 */
:host {
  /* Inherited properties cross the shadow boundary, so a host page's
     text-transform, letter-spacing or font would otherwise restyle the UI. */
  font: 400 16px/1.5 system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-variant: normal;
  color: initial;
  letter-spacing: normal;
  word-spacing: normal;
  text-transform: none;
  text-align: start;
  text-indent: 0;
  text-shadow: none;
  white-space: normal;
  direction: ltr;
  writing-mode: horizontal-tb;
  hyphens: manual;
  word-break: normal;
  overflow-wrap: normal;
  tab-size: 8;
  list-style: none;
  quotes: initial;
  caret-color: auto;
  user-select: auto;
  -webkit-text-fill-color: initial;
  -webkit-text-stroke: 0;
  text-rendering: auto;
  -webkit-text-size-adjust: 100%; /* 2 */
  -webkit-tap-highlight-color: transparent; /* 3 */
  font-feature-settings: "kern"; /* 4 */
  -webkit-font-feature-settings: "kern"; /* 5 */
  -moz-font-feature-settings: "kern"; /* 5 */
  -webkit-font-smoothing: antialiased; /* 5 */
  -moz-osx-font-smoothing: grayscale; /* 5 */
}

/* Vertical rhythm */
/* ============================================ */
p,
table,
blockquote,
address,
pre,
iframe,
form,
figure,
dl {
  margin: 0;
}

/* Headings */
/* ============================================ */
h1,
h2,
h3,
h4,
h5,
h6 {
  margin: 0;
  font-size: inherit;
  font-weight: inherit;
}

/* Lists (enumeration) */
/* ============================================ */
ul,
ol,
menu {
  list-style: none;
  margin: 0;
  padding: 0;
}

/* Lists (definition) */
/* ============================================ */
dd {
  margin-left: 0;
}

/* Grouping content */
/* ============================================ */
/**
 * 1. Add the correct box sizing in Firefox.
 * 2. Show the overflow in Edge and IE.
 */
hr {
  clear: both;
  margin: 0;
  border-top-width: 1px;
  height: 0; /* 1 */
  box-sizing: content-box; /* 1 */
  overflow: visible; /* 2 */
  color: inherit;
}

/**
 * 1. Correct the inheritance and scaling of font size in all browsers.
 * 2. Correct the odd \`em\` font sizing in all browsers.
 * 3. Wrap lines by default instead of overflow.
 */
pre {
  font-family: inherit; /* 1 */
  font-size: inherit; /* 2 */
  white-space: pre-line; /* 3 */
}

address {
  font-style: inherit;
}

/* Text-level semantics */
/* ============================================ */
/**
 * Remove the gray background on active links in IE 10.
 */
a {
  background-color: transparent;
  text-decoration: none;
  color: inherit;
}

/**
 * 1. Remove the bottom border in Chrome 57-
 * 2. Add the correct text decoration in Chrome, Edge, IE, Opera, and Safari.
 */
abbr[title] {
  border-bottom: none; /* 1 */
  text-decoration: none; /* 2 */
}

/**
 * Add the correct font weight in Chrome, Edge, and Safari.
 */
b,
strong {
  font-weight: bolder;
}

/**
 * 1. Correct the inheritance and scaling of font size in all browsers.
 * 2. Correct the odd \`em\` font sizing in all browsers.
 */
code,
kbd,
samp {
  font-family: "Menlo", "Monaco", "Consolas", "Courier New", monospace; /* 1 */
  font-size: inherit; /* 2 */
}

/**
 * Add the correct font size in all browsers.
 */
small {
  font-size: 80%;
}

/**
 * Prevent \`sub\` and \`sup\` elements from affecting the line height in all browsers.
 */
sub,
sup {
  position: relative;
  vertical-align: baseline;
  line-height: 0;
  font-size: 75%;
}

sub {
  bottom: -0.25em;
}

sup {
  top: -0.5em;
}

/* Replaced content */
/* ============================================ */
/**
 * Prevent vertical alignment issues.
 */
svg,
img,
embed,
object,
iframe {
  vertical-align: bottom;
}

/*
 * 1. Remove image default bottom space.
 * 2. Prevent image from overflowing the container.
 */
img {
  display: block;
  max-width: 100%;
}

/**
 * Prevent alignment issues on Safari.
 */
@supports (background: -webkit-named-image(i)) {
  svg {
    will-change: transform;
  }
}
/* Forms */
/* ============================================ */
/**
 * Reset form fields to make them styleable.
 * 1. Make form elements stylable across systems iOS especially.
 * 2. Inherit text-transform from parent.
 */
button,
input,
optgroup,
select,
textarea {
  -webkit-appearance: none; /* 1 */
  appearance: none;
  border-radius: 0;
  margin: 0;
  padding: 0;
  background: transparent;
  vertical-align: middle;
  text-align: inherit;
  text-transform: inherit; /* 2 */
  font: inherit;
  color: inherit;
}

/**
 * Correct cursors for clickable elements.
 */
button,
[type=button],
[type=reset],
[type=submit] {
  cursor: pointer;
}

button:disabled,
[type=button]:disabled,
[type=reset]:disabled,
[type=submit]:disabled {
  cursor: default;
}

/**
 * Clickable labels and selects.
 */
select,
label {
  cursor: pointer;
}

/**
 * Improve outlines for Firefox and unify style with input elements & buttons.
 */
:-moz-focusring {
  outline: auto;
}

select:disabled {
  opacity: inherit;
}

/**
 * 1. Remove padding.
 */
option {
  padding: 0; /* 1 */
}

/**
 * Reset to invisible
 */
fieldset {
  margin: 0;
  padding: 0;
  min-width: 0;
}

legend {
  display: contents;
  padding: 0;
}

/**
 * Add the correct vertical alignment in Chrome, Firefox, and Opera.
 */
progress {
  vertical-align: baseline;
}

/**
 * Remove the default vertical scrollbar in IE 10+.
 */
textarea {
  overflow: auto;
}

/**
 * Remove increment and decrement buttons in Chrome.
 */
[type=number]::-webkit-inner-spin-button,
[type=number]::-webkit-outer-spin-button {
  -webkit-appearance: none;
}

/**
 * Correct the outline style in Safari.
 */
[type=search] {
  outline-offset: -2px;
}

/**
 * Remove the inner padding in Chrome and Safari on macOS.
 */
[type=search]::-webkit-search-decoration {
  -webkit-appearance: none;
}

/*
 * Remove the ‘X’ from Chrome and Safari.
 */
[type=search]::-webkit-search-decoration,
[type=search]::-webkit-search-cancel-button,
[type=search]::-webkit-search-results-button,
[type=search]::-webkit-search-results-decoration {
  display: none;
}

/**
 * 1. Hide file input completely.
 * 2. Remove selected file text.
 * 3. Set cursor to pointer for all browsers.
 */
[type=file] {
  opacity: 0; /* 1 */
  font-size: 0; /* 2 */
  cursor: pointer; /* 3 */
}

/**
	* Fix appearance for Firefox
	*/
[type=number] {
  -moz-appearance: textfield;
}

/**
 * Set cursor to pointer for all browsers.
 */
[type=range] {
  cursor: pointer;
}

/**
 * Reset slider thumbs to make them styleable.
 */
[type=range]::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
}

[type=range]::-moz-range-thumb {
  -moz-appearance: none;
  appearance: none;
  border-width: 0;
  border-radius: 0;
  background-color: transparent;
}

/* Interactive */
/* ============================================ */
/*
 * Add the correct display in Edge, IE 10+, and Firefox.
 */
details {
  display: block;
}

/*
 * Add the correct display in all browsers.
 */
summary {
  display: list-item;
}

/*
 * Remove outline for editable content.
 */
[contenteditable]:focus {
  outline: auto;
}

/* Tables */
/* ============================================ */
/**
1. Correct table border color inheritance in all Chrome and Safari.
*/
table {
  border-color: inherit; /* 1 */
  border-collapse: collapse;
}

caption {
  text-align: left;
}

td,
th {
  vertical-align: top;
  padding: 0;
}

th {
  text-align: left;
  font-weight: inherit;
}

/* Misc */
/* ============================================ */
/*
 * Make placeholder style consistent across all browsers.
 */
::placeholder {
  color: #999;
  opacity: 1;
}

/*
 * Hide focus outline but keep it visible for Windows High Contrast Mode.
 */
:focus {
  outline-style: solid;
  outline-color: transparent;
}

/*
 * Hide input arrow when used with datalist.
 */
::-webkit-calendar-picker-indicator {
  display: none !important;
}`,fw=[dw,Tv,Zp,sw,G5,Ev,sd,lw,ew,rw,iw,ow].join(`
`);function Wu(s,r="filtered",c){const{name:u,path:f}=qr(s,c);if(r==="off")return{name:u,elementName:u,path:f,reactComponents:null};const d=U2(s,{mode:r});return{name:d.path?`${d.path} ${u}`:u,elementName:u,path:f,reactComponents:d.path}}var dp=!1,C_={outputDetail:"standard",autoClearAfterCopy:!1,annotationColorId:"blue",blockInteractions:!0,reactEnabled:!0,markerClickBehavior:"edit",webhookUrl:"",webhooksEnabled:!0},fp=s=>{if(!s||!s.trim())return!1;try{const r=new URL(s.trim());return r.protocol==="http:"||r.protocol==="https:"}catch{return!1}},Js=s=>s.metaKey||s.ctrlKey,dc=[{id:"indigo",label:"Indigo",srgb:"#6155F5",p3:"color(display-p3 0.38 0.33 0.96)"},{id:"blue",label:"Blue",srgb:"#0088FF",p3:"color(display-p3 0.00 0.53 1.00)"},{id:"cyan",label:"Cyan",srgb:"#00C3D0",p3:"color(display-p3 0.00 0.76 0.82)"},{id:"green",label:"Green",srgb:"#34C759",p3:"color(display-p3 0.20 0.78 0.35)"},{id:"yellow",label:"Yellow",srgb:"#FFCC00",p3:"color(display-p3 1.00 0.80 0.00)"},{id:"orange",label:"Orange",srgb:"#FF8D28",p3:"color(display-p3 1.00 0.55 0.16)"},{id:"red",label:"Red",srgb:"#FF383C",p3:"color(display-p3 1.00 0.22 0.24)"}],_w=[...dc.map(s=>`
    [data-agentation-accent="${s.id}"] {
      --agentation-color-accent: ${s.srgb};
    }
    @supports (color: color(display-p3 0 0 0)) {
      [data-agentation-accent="${s.id}"] {
        --agentation-color-accent: ${s.p3};
      }
    }
  `),`:host {
    ${dc.map(s=>`--agentation-color-${s.id}: ${s.srgb};`).join(`
`)}
  }`,`@supports (color: color(display-p3 0 0 0)) {
    :host {
      ${dc.map(s=>`--agentation-color-${s.id}: ${s.p3};`).join(`
`)}
    }
  }`].join("");function N_(s){let r=s;for(let u=Xl(r.ownerDocument);u;u=Xl(r.ownerDocument))r=u;let c=r;for(;c&&c!==document.body;){const f=window.getComputedStyle(c).position;if(f==="fixed"||f==="sticky")return!0;c=c.parentElement}return!1}function rc(s){return s.kind!=="placement"&&s.kind!=="rearrange"&&s.status!=="resolved"&&s.status!=="dismissed"}function Ku(s){const r=D_(s),c=r.found?r:K2(s);if(c.found&&c.source)return W2(c.source,"path")}function mw(s={}){const r=cv(s.useHashLocation??!1),c=h.useState(!1),u=fv(s.portalContainer);return u?L_.createPortal(h.createElement(hw,{...s,key:s.useHashLocation?r:void 0,pathname:r,activeState:c,portalHost:u}),u):null}function hw({pathname:s,activeState:r,portalHost:c,useHashLocation:u=!1,appName:f,enableKeyboardShortcuts:d=!0,identifyingAttributes:m=V_,copyFormat:g="markdown",onOpenSource:p,portalContainer:j,demoAnnotations:k,demoDelay:b=1e3,enableDemoMode:v=!1,onAnnotationAdd:z,onAnnotationDelete:E,onAnnotationUpdate:$,onAnnotationsClear:B,onCopy:F,onSubmit:Me,copyToClipboard:te=!0,endpoint:C,sessionId:ie,onSessionCreated:de,webhookUrl:W,className:Ee}){var Co,pi,yi,xi,Nc,Mc,Ec;const[ne,Ce]=r,[,pe]=h.useState(0),[V]=h.useState(()=>J1(document,()=>pe(x=>x+1)));h.useEffect(()=>(V.start(),()=>V.stop()),[V]);const Ie=typeof g=="object"?g.attribute:void 0,he=h.useMemo(()=>Ie?[...m,Ie]:m,[m,Ie]),ue=h.useRef(!0);h.useLayoutEffect(()=>(ue.current=!0,()=>{ue.current=!1}),[]);const G=h.useCallback(x=>u?uv(JSON.stringify([C,s]),x):x(),[C,s,u]),we=h.useRef(new Map),fe=h.useRef(new Set),Oe=x=>rc(x)&&!fe.current.has(x.id),Se=async(...x)=>{const S=await tp(...x),N=x[2],T=N.id;if(T&&(we.current.set(T,S.id),fe.current.has(T)&&fe.current.add(S.id)),!u){const Q=new URL(N.url||window.location.href).pathname,H=Zs(Q).find(Y=>Y.id===T);try{if(fe.current.has(T))await Gu(x[0],S.id);else if(H&&H.comment!==N.comment)return await w_(x[0],S.id,{comment:H.comment}),{...S,comment:H.comment}}catch(Y){console.warn("[Agentation] Failed to apply changes made during sync:",Y)}}return S},Qe=x=>u?{...x,annotations:x.annotations.filter(S=>!fe.current.has(S.id)&&dv(S.url||x.url,s,window.location.origin))}:x,Ut=h.useRef(s);Ut.current=s;const Ye=(x,S,N,T=s)=>{const Q=G1(x,Zs(T),S,we.current).filter(Oe);T===Ut.current&&ue.current&&Xt(Q),x_(T,Q,N)},[Ne,Xt]=h.useState([]),[xe,We]=h.useState(!0),[gn,Qt]=h.useState(()=>k2()),[mn,Re]=h.useState(!1);h.useLayoutEffect(()=>{Xp()},[]);const Ue=h.useRef(null),un=h.useRef(null),ul=h.useRef(null),Cn=h.useRef(null),bn=h.useRef(!1),re=h.useRef(!1),ye=h.useRef(!1);h.useLayoutEffect(()=>{var x,S,N;ne&&bn.current?(bn.current=!1,(S=(x=ul.current)==null?void 0:x.querySelector("button:not(:disabled)"))==null||S.focus()):!ne&&re.current&&(re.current=!1,(N=un.current)==null||N.focus())},[ne]),h.useEffect(()=>{const x=N=>{const T=Ue.current;T&&N.composedPath().includes(T)&&N.stopPropagation()},S=["mousedown","click","pointerdown"];return S.forEach(N=>c.addEventListener(N,x)),()=>{S.forEach(N=>c.removeEventListener(N,x))}},[c]);const[Ge,nt]=h.useState(!1),[lt,rt]=h.useState(!1),[it,Fe]=h.useState(null),[Nt,an]=h.useState({x:0,y:0}),[oe,O]=h.useState(null),[I,P]=h.useState(!1),ae=D0(),Be=D0(),[ke,_e]=h.useState("idle"),[at,ot]=h.useState(!1),ct=h.useRef(new Set),Ke=h.useRef(new Set),Ze=h.useRef(),bt=h.useCallback(()=>{!ct.current.size&&!Ze.current&&ot(!1)},[]);h.useEffect(()=>()=>clearTimeout(Ze.current),[]);const[Bt,pn]=h.useState(null),[pt,At]=h.useState(null),[kt,ht]=h.useState([]),[Nn,vt]=h.useState(null),Je=h.useRef(null);h.useEffect(()=>()=>{Je.current&&clearTimeout(Je.current)},[]);const[je,Gt]=h.useState(null),vn=h.useRef(null),ho=h.useRef(!1),[Ql,yn]=h.useState(!1);h.useLayoutEffect(()=>{if(je||!vn.current)return;const x=vn.current;if(vn.current=null,oe)return;const S=ne&&x.isConnected&&!x.disabled?x:un.current;S==null||S.focus({preventScroll:!0})},[je,ne,oe]);const[Fn,Hn]=h.useState(null),[Mn,dl]=h.useState([]),[sl,Gr]=h.useState(0),[hc,Vr]=h.useState(!1),[Ot,Ps]=h.useState(!1),[vl,Fr]=h.useState(!1),[Yn,go]=h.useState(!1),[Wr,$a]=h.useState("main"),[Kr,Ua]=h.useState(!1),[ut,En]=h.useState(!1),[wn,Gl]=h.useState(!1),[ft,eo]=h.useState([]),[is,po]=h.useState(null),Uo=h.useRef(!1),[Yt,ea]=h.useState(!1),[yo,xo]=h.useState(!1),[cs,us]=h.useState(1),[Tn,ad]=h.useState("new-page"),[qn,ta]=h.useState(""),[Ho,ds]=h.useState(!1),[be,Rl]=h.useState(null),Ha=h.useRef(!1),Zr=h.useRef({rearrange:null,placements:[]}),yt=h.useRef({rearrange:null,placements:[]}),[gc,na]=h.useState(0),[Ya,to]=h.useState(0),[On,Jr]=h.useState([]),[fs,qa]=h.useState(null),Pr=h.useRef({designPlacements:ft,rearrangeState:be,blankCanvas:Yt,wireframePurpose:qn});Pr.current={designPlacements:ft,rearrangeState:be,blankCanvas:Yt,wireframePurpose:qn};const _s=h.useRef({placements:On,rearrange:fs}),Xn=h.useRef(new Set),ms=h.useRef(new Set),wl=h.useRef(null),bo=h.useRef(),Yo=ut&&ne&&!wn&&Yt;h.useEffect(()=>{if(Yo){xo(!1);const x=nd(()=>{xo(!0)});return()=>cancelAnimationFrame(x)}else xo(!1)},[Yo]);const Ia=h.useRef(new Map),Xa=h.useRef([]),vo=h.useRef(new Map),qo=h.useRef(null),[fl,Qa]=h.useState(!1),[_l,rd]=h.useState([]),ei=h.useRef(_l);ei.current=_l;const[la,no]=h.useState(null),oa=h.useRef(null);h.useRef(!1),h.useRef([]),h.useRef(0),h.useRef(null),h.useRef(null),h.useRef(1);const[sa,Io]=h.useState(!1),zl=h.useRef(null),[Wn,Vl]=h.useState([]),hs=h.useRef(!1),Sn=()=>{Ua(!0)},lo=()=>{Ua(!1)},Ga=()=>{sa||(zl.current=Ht(()=>Io(!0),850))},gs=()=>{zl.current&&(clearTimeout(zl.current),zl.current=null),Io(!1),lo()};h.useEffect(()=>()=>{zl.current&&clearTimeout(zl.current)},[]);const[wt,ti]=h.useState(()=>{try{const x=JSON.parse(localStorage.getItem("feedback-toolbar-settings")??"");return{...C_,...x,annotationColorId:dc.find(S=>S.id===x.annotationColorId)?x.annotationColorId:C_.annotationColorId}}catch{return C_}}),[ml,ni]=h.useState(!0),[aa,ra]=h.useState(!1),pc=h.useCallback(x=>{ti(S=>({...S,...x}))},[]),Kn=h.useCallback(()=>{var x;(x=Ue.current)==null||x.classList.add(le.disableTransitions),ni(S=>!S),nd(()=>{var S;(S=Ue.current)==null||S.classList.remove(le.disableTransitions)})},[]),wo=!1,Sl="off",[dn,li]=h.useState(u?null:ie??null),Va=h.useRef(!1),[oo,Al]=h.useState(C?"connecting":"disconnected"),[ln,Fa]=h.useState(null),[Wa,yc]=h.useState(!1),ia=h.useRef(null),ps=h.useRef(!1),Xo=h.useRef(new Set),Ka=h.useRef(new Map),xc=h.useCallback(x=>{Xo.current.add(x),Ol.current===x&&(Ol.current=null)},[]),[So,Za]=h.useState(new Set),[hl,ca]=h.useState(!1),[Qo,ys]=h.useState(!1),[so,ua]=h.useState(!1),Go=h.useRef(null),jl=h.useRef(null),da=h.useRef(null),Vo=h.useRef(null),fa=h.useRef(!1),bc=h.useRef(0),Ol=h.useRef(null),vc=h.useRef(null),oi=8,id=50,si=h.useRef(null),Ja=h.useRef(null),_a=h.useRef(null),cd=h.useCallback(()=>$a("main"),[]);h.useEffect(()=>{Yn||Ua(!1)},[Yn]),h.useLayoutEffect(()=>{var x,S;Yn&&ye.current&&(ye.current=!1,(S=(x=Ue.current)==null?void 0:x.querySelector("[data-agentation-settings-panel] button"))==null||S.focus())},[Yn]);const Pa=ne&&xe&&!ut;h.useEffect(()=>{if(Pa)rt(!1),nt(!0),Xo.current.clear();else if(Ge){rt(!0);const x=Ht(()=>{nt(!1),rt(!1)},250);return()=>clearTimeout(x)}},[Pa]),h.useEffect(()=>{Ps(!0),Gr(window.scrollY);const x=Zs(s);Xt(x.filter(rc)),dp||(ra(!0),dp=!0,Ht(()=>ra(!1),750));try{const S=localStorage.getItem("feedback-toolbar-theme");S!==null&&ni(S==="dark")}catch{}try{const S=localStorage.getItem("feedback-toolbar-position");if(S){const N=JSON.parse(S);typeof N.x=="number"&&typeof N.y=="number"&&Fa(N)}}catch{}},[s]),h.useEffect(()=>{Ot&&localStorage.setItem("feedback-toolbar-settings",JSON.stringify(wt))},[wt,Ot]),h.useEffect(()=>{Ot&&localStorage.setItem("feedback-toolbar-theme",ml?"dark":"light")},[ml,Ot]);const er=h.useRef(!1);h.useEffect(()=>{const x=er.current;er.current=Wa,x&&!Wa&&ln&&Ot&&localStorage.setItem("feedback-toolbar-position",JSON.stringify(ln))},[Wa,ln,Ot]),h.useEffect(()=>{if(!C||!Ot||Va.current)return;Va.current=!0,Al("connecting");const x=window.location.href;G(async()=>{try{const N=S2(s),T=ie||N;let Q=!1;if(T)try{const H=Zs(s),Y=Qe(await ep(C,T));Xa.current=Y.annotations.filter(Te=>Te.kind==="placement"||Te.kind==="rearrange"),ue.current&&(li(Y.id),Al("connected")),b_(s,Y.id),Q=!0;const ce=Zs(s).filter(rc),me=new Set(Y.annotations.map(Te=>Te.id)),ge=ce.filter(Te=>!me.has(Te.id));if(ge.length>0){const qe=`${typeof window<"u"?window.location.origin:""}${s}`,dt=(await Promise.allSettled(ge.map(Ve=>Se(C,Y.id,{...Ve,sessionId:Y.id,url:qe})))).map((Ve,st)=>Ve.status==="fulfilled"?Ve.value:(console.warn("[Agentation] Failed to sync annotation:",Ve.reason),ge[st])),$e=[...Y.annotations,...dt];Ye(H,$e,Y.id)}else Ye(H,Y.annotations,Y.id)}catch(H){console.warn("[Agentation] Could not join session, creating new:",H),j2(s)}if(!Q){const H=await v_(C,x);b_(s,H.id),ue.current&&(li(H.id),Al("connected"),de==null||de(H.id));const Y=u?new Map([[s,Zs(s)]]):h2(),ce=typeof window<"u"?window.location.origin:"",me=[];for(const[ge,Te]of Y){const qe=Te.filter($e=>rc($e)&&!$e._syncedTo);if(qe.length===0)continue;const Pe=`${ce}${ge}`,dt=ge===s;me.push((async()=>{try{const $e=dt?H:await v_(C,Pe),st=(await Promise.allSettled(qe.map(St=>Se(C,$e.id,{...St,sessionId:$e.id,url:Pe})))).map((St,$t)=>St.status==="fulfilled"?St.value:(console.warn("[Agentation] Failed to sync annotation:",St.reason),qe[$t]));Ye(qe,st,$e.id,ge)}catch($e){console.warn(`[Agentation] Failed to sync annotations for ${ge}:`,$e)}})())}await Promise.allSettled(me)}}catch(N){ue.current&&Al("disconnected"),console.warn("[Agentation] Failed to initialize session, using local storage:",N)}})},[C,ie,Ot,de,s,G]),h.useEffect(()=>{if(!C||!Ot)return;const x=async()=>{try{(await fetch(`${C}/health`)).ok?Al("connected"):Al("disconnected")}catch{Al("disconnected")}};x();const S=Qp(x,1e4);return()=>clearInterval(S)},[C,Ot]);const jo=h.useRef(Ne),wc=h.useRef(!1);h.useLayoutEffect(()=>{jo.current=Ne,wc.current=Ne.length>0||ft.length>0||((be==null?void 0:be.sections.length)??0)>0},[Ne,ft.length,be==null?void 0:be.sections.length]);const tr=h.useCallback(x=>{const S=ct.current.has(x);if(S&&(Ke.current.delete(x),Ke.current.size))return;const N=S?new Set(ct.current):new Set([x]);S&&(ct.current.clear(),bt());for(const H of N)Ka.current.delete(H),Xo.current.delete(H);Xt(H=>H.filter(Y=>!N.has(Y.id))),Za(H=>new Set([...H].filter(Y=>!N.has(Y))));const T=S?[]:jo.current.filter(H=>H.kind!=="placement"&&H.kind!=="rearrange"),Q=T.findIndex(H=>H.id===x);Q>=0&&Q<T.length-1&&(vt(H=>H===null?Q:Math.min(H,Q)),Je.current&&clearTimeout(Je.current),Je.current=Ht(()=>vt(null),200))},[bt]);h.useEffect(()=>!C||!Ot||!dn?void 0:P2(C,dn,()=>wc.current,N=>{var H,Y;const{id:T,kind:Q}=N;if(Q==="placement"){for(const[ce,me]of Ia.current)if(me===T){(H=qo.current)==null||H.placements.forget(ce),eo(ge=>ge.filter(Te=>Te.id!==ce));break}}else if(Q==="rearrange"){for(const[ce,me]of vo.current)if(me===T){(Y=qo.current)==null||Y.rearrange.forget(ce),Rl(ge=>{if(!ge)return null;const Te=ge.sections.filter(qe=>qe.id!==ce);return Te.length===0?null:{...ge,sections:Te}});break}}else{if(!jo.current.some(ce=>ce.id===T))return;Za(ce=>new Set(ce).add(T))}}),[C,Ot,dn]),h.useEffect(()=>{if(!C||!Ot)return;const x=vc.current==="disconnected",S=oo==="connected";vc.current=oo,x&&S&&G(async()=>{try{const T=Zs(s).filter(rc);if(T.length===0)return;const H=`${typeof window<"u"?window.location.origin:""}${s}`;let Y=dn,ce=[];if(Y)try{ce=Qe(await ep(C,Y)).annotations}catch{Y=null}Y||(Y=(await v_(C,H)).id,ue.current&&li(Y),b_(s,Y));const me=new Set(ce.map(Te=>Te.id)),ge=T.filter(Te=>!me.has(Te.id));if(ge.length>0){const qe=(await Promise.allSettled(ge.map(dt=>Se(C,Y,{...dt,sessionId:Y,url:H})))).map((dt,$e)=>dt.status==="fulfilled"?dt.value:(console.warn("[Agentation] Failed to sync annotation on reconnect:",dt.reason),ge[$e])),Pe=[...ce,...qe];Ye(T,Pe,Y)}}catch(T){console.warn("[Agentation] Failed to sync on reconnect:",T)}})},[oo,C,Ot,dn,s,G]);const Sc=h.useCallback(()=>{mn||(Re(!0),go(!1),Ce(!1),Ht(()=>{C2(!0),Qt(!0),Re(!1)},400))},[mn]);h.useEffect(()=>{if(!v||!Ot||!k||k.length===0||Ne.length>0)return;const x=[];return x.push(Ht(()=>{Ce(!0)},b-200)),k.forEach((S,N)=>{const T=b+N*300;x.push(Ht(()=>{const Q=document.querySelector(S.selector);if(!Q)return;const H=Un(Q),{name:Y,path:ce}=qr(Q),me={id:`demo-${Date.now()}-${N}`,x:(H.left+H.width/2)/window.innerWidth*100,y:H.top+H.height/2+window.scrollY,comment:S.comment,element:Y,elementPath:ce,timestamp:Date.now(),selectedText:S.selectedText,boundingBox:{x:H.left,y:H.top+window.scrollY,width:H.width,height:H.height},nearbyText:nc(Q),cssClasses:lc(Q)};Xt(ge=>[...ge,me])},T))}),()=>{x.forEach(clearTimeout)}},[v,Ot,k,b]),h.useEffect(()=>{const x=()=>{Gr(window.scrollY),pe(S=>S+1),Vr(!0),_a.current&&clearTimeout(_a.current),_a.current=Ht(()=>{Vr(!1)},150)};return V.addEventListener("scroll",x,{passive:!0,capture:!0}),()=>{V.removeEventListener("scroll",x,!0),_a.current&&clearTimeout(_a.current)}},[V]),h.useEffect(()=>{if(!Ot)return;const x=Ne.filter(S=>!So.has(S.id));x.length>0?dn?x_(s,x,dn):dy(s,x):localStorage.removeItem(Z_(s))},[Ne,s,Ot,dn,at,So]),h.useEffect(()=>{if(Ot&&!Uo.current){Uo.current=!0;const x=g2(s);x.length>0&&eo(x)}},[Ot,s]),h.useEffect(()=>{if(Ot&&Uo.current&&!Yt){const x=ft.filter(S=>!On.includes(S));x.length>0?p2(s,x):y2(s)}},[ft,s,Ot,Yt,On]),h.useEffect(()=>{if(Ot&&!Ha.current){Ha.current=!0;const x=x2(s);if(x){const S={...x,sections:x.sections.map(N=>({...N,currentRect:N.currentRect??{...N.originalRect}}))};Rl(S)}}},[Ot,s]),h.useEffect(()=>{Ot&&Ha.current&&!Yt&&(be&&be!==fs?b2(s,be):v2(s))},[be,s,Ot,Yt,fs]);const nr=h.useRef(!1);h.useEffect(()=>{if(Ot&&!nr.current){nr.current=!0;const x=w2(s);x&&(yt.current={rearrange:x.rearrange,placements:x.placements||[]},x.purpose&&ta(x.purpose))}},[Ot,s]),h.useEffect(()=>{var S,N,T;if(!Ot||!nr.current||at)return;const x=yt.current;Yt?(((S=be==null?void 0:be.sections)==null?void 0:S.length)??0)>0||ft.length>0||qn?P0(s,{rearrange:be,placements:ft,purpose:qn}):Qu(s):(((T=(N=x.rearrange)==null?void 0:N.sections)==null?void 0:T.length)??0)>0||x.placements.length>0||qn?P0(s,{rearrange:x.rearrange,placements:x.placements,purpose:qn}):Qu(s)},[be,ft,qn,Yt,s,Ot,at]),h.useEffect(()=>{ut&&!be&&Rl({sections:[],originalOrder:[],detectedAt:Date.now()})},[ut,be]),h.useEffect(()=>{if(!C||!dn)return;const x={create:N=>G(()=>tp(C,dn,N)),update:(N,T)=>G(()=>w_(C,N,T)),remove:N=>G(()=>Gu(C,N))};Ia.current=new Map,vo.current=new Map;const S={placements:rp(x,Ia.current,Xa.current.filter(N=>N.kind==="placement")),rearrange:rp(x,vo.current,Xa.current.filter(N=>N.kind==="rearrange"))};return qo.current=S,()=>{S.placements.dispose(),S.rearrange.dispose(),qo.current===S&&(qo.current=null)}},[C,dn,s,G]),h.useEffect(()=>{var S;const x=window.location.pathname+window.location.search+window.location.hash;(S=qo.current)==null||S.placements.replace(ft.filter(N=>!On.includes(N)).map(N=>({id:N.id,x:N.x/window.innerWidth*100,y:N.y,comment:`Place ${N.type} at (${Math.round(N.x)}, ${Math.round(N.y)}), ${N.width}×${N.height}px${N.text?` — "${N.text}"`:""}`,element:`[design:${N.type}]`,elementPath:"[placement]",timestamp:N.timestamp,url:x,intent:"change",severity:"important",kind:"placement",placement:{componentType:N.type,width:N.width,height:N.height,scrollY:N.scrollY,text:N.text}})))},[ft,C,dn,s,On]),h.useEffect(()=>{const x=qo.current;if(!x)return;if(be===fs){x.rearrange.replace([]);return}const S=Ht(()=>{const N=window.location.pathname+window.location.search+window.location.hash,T=[];for(const Q of(be==null?void 0:be.sections)??[]){const H=Q.originalRect,Y=Q.currentRect,ce=Math.abs(H.x-Y.x)>1||Math.abs(H.y-Y.y)>1||Math.abs(H.width-Y.width)>1||Math.abs(H.height-Y.height)>1;if(!ce&&!Q.note)continue;const me=Q.note?` — "${Q.note}"`:"";T.push({id:Q.id,x:Y.x/window.innerWidth*100,y:Y.y,comment:ce?`Move ${Q.label} section (${Q.tagName}) — from (${Math.round(H.x)},${Math.round(H.y)}) ${Math.round(H.width)}×${Math.round(H.height)} to (${Math.round(Y.x)},${Math.round(Y.y)}) ${Math.round(Y.width)}×${Math.round(Y.height)}${me}`:`Note on ${Q.label} section (${Q.tagName})${me}`,element:Q.selector,elementPath:"[rearrange]",timestamp:be.detectedAt,url:N,intent:"change",severity:"important",kind:"rearrange",rearrange:{selector:Q.selector,label:Q.label,tagName:Q.tagName,originalRect:H,currentRect:Y}})}x.rearrange.replace(T)},300);return()=>clearTimeout(S)},[be,C,dn,s,fs]);const lr=h.useCallback(()=>{clearTimeout(bo.current),Gl(!1),En(!0)},[]);h.useEffect(()=>()=>clearTimeout(bo.current),[]);const xs=h.useCallback(()=>{Gl(!0),En(!1),po(null),clearTimeout(bo.current),bo.current=Ht(()=>{Gl(!1)},300)},[]),ao=h.useCallback(()=>{var S,N,T;const x=(S=un.current)==null?void 0:S.getRootNode();re.current=!!(x!=null&&x.activeElement)&&!!((N=Ue.current)!=null&&N.contains(x.activeElement)),re.current&&((T=x==null?void 0:x.activeElement)==null||T.blur()),go(!1),ut&&(Gl(!0),En(!1),po(null),clearTimeout(bo.current),bo.current=Ht(()=>{Gl(!1)},300)),Ce(!1)},[ut]),jc=h.useCallback(()=>{vl||(Q1(),Fr(!0))},[vl]),or=h.useCallback(()=>{vl&&(O0(),Fr(!1))},[vl]),ai=h.useCallback(()=>{vl?or():jc()},[vl,jc,or]),bs=h.useCallback((x=Wn)=>{const S=x.filter(Y=>Y.element.isConnected);if(S.length===0){Vl([]);return}const N=S[0],T=N.element,Q=S.length>1,H=S.map(Y=>Un(Y.element));if(Q){const Y={left:Math.min(...H.map(st=>st.left)),top:Math.min(...H.map(st=>st.top)),right:Math.max(...H.map(st=>st.right)),bottom:Math.max(...H.map(st=>st.bottom))},ce=S.slice(0,5).map(st=>st.name).join(", "),me=S.length>5?` +${S.length-5} more`:"",ge=H.map(st=>({x:st.left,y:st.top+window.scrollY,width:st.width,height:st.height})),qe=S[S.length-1].element,Pe=H[H.length-1],dt=Pe.left+Pe.width/2,$e=Pe.top+Pe.height/2,Ve=N_(qe);O({id:Date.now().toString(),x:dt/window.innerWidth*100,y:Ve?$e:$e+window.scrollY,clientY:$e,element:`${S.length} elements: ${ce}${me}`,elementPath:"multi-select",boundingBox:{x:Y.left,y:Y.top+window.scrollY,width:Y.right-Y.left,height:Y.bottom-Y.top},isMultiSelect:!0,isFixed:Ve,elementBoundingBoxes:ge,multiSelectElements:S.map(st=>st.element),targetElement:qe,fullPath:cc(T),accessibility:Uu(T),computedStyles:$u(T),computedStylesObj:Bu(T),nearbyElements:Lu(T),cssClasses:lc(T),nearbyText:nc(T),sourceFile:Ku(T),attributes:ic(T,he)})}else{const Y=H[0],ce=N_(T);O({id:Date.now().toString(),x:Y.left/window.innerWidth*100,y:ce?Y.top:Y.top+window.scrollY,clientY:Y.top,element:N.name,elementPath:N.path,boundingBox:{x:Y.left,y:ce?Y.top:Y.top+window.scrollY,width:Y.width,height:Y.height},isFixed:ce,fullPath:cc(T),accessibility:Uu(T),computedStyles:$u(T),computedStylesObj:Bu(T),nearbyElements:Lu(T),cssClasses:lc(T),nearbyText:nc(T),reactComponents:N.reactComponents,targetElement:T,sourceFile:Ku(T),attributes:ic(T,he)})}Vl([]),Fe(null)},[Wn,he]);h.useEffect(()=>{ne||(O(null),Gt(null),Hn(null),dl([]),Fe(null),go(!1),Vl([]),hs.current=!1,vl&&or())},[ne,vl,or]),h.useEffect(()=>()=>{O0()},[]),h.useEffect(()=>{if(!ne)return;const x=["p","span","h1","h2","h3","h4","h5","h6","li","td","th","label","blockquote","figcaption","caption","legend","dt","dd","pre","code","em","strong","b","i","u","s","a","time","address","cite","q","abbr","dfn","mark","small","sub","sup","[contenteditable]"].join(", "),S=document.createElement("style");return S.id="agentation-cursor",S.textContent=`
      body { cursor: crosshair !important; }
      body :is(${x}) { cursor: text !important; }
    `,document.head.appendChild(S),()=>{const N=document.getElementById("agentation-cursor");N&&N.remove()}},[ne]),h.useEffect(()=>{if(la!==null&&ne)return document.documentElement.setAttribute("data-drawing-hover",""),()=>document.documentElement.removeAttribute("data-drawing-hover")},[la,ne]),h.useEffect(()=>{if(!ne||oe||je||fl||ut)return;let x=null;const S=(H,Y,ce)=>{const me=ed(H,Y),ge=ce?B0(H,Y):me;if(!ge||ol(ge,"[data-feedback-toolbar], [data-annotation-popup], [data-annotation-marker]")){Fe(null);return}const{name:Te,elementName:qe,path:Pe,reactComponents:dt}=Wu(ge,Sl,he);Fe({element:Te,elementName:qe,elementPath:Pe,rect:Un(ge),reactComponents:dt,isPiercing:ce&&ge!==me}),an({x:H,y:Y})},N=H=>{const Y=H.composedPath()[0]||H.target;if(ol(Y,"[data-feedback-toolbar], [data-annotation-popup], [data-annotation-marker]")){x=null,Fe(null);return}x={x:H.clientX,y:H.clientY},S(H.clientX,H.clientY,Js(H))},T=H=>{(H.key==="Meta"||H.key==="Control")&&x&&S(x.x,x.y,Js(H))},Q=()=>{x=null,Fe(null)};return V.addEventListener("mousemove",N),V.addEventListener("keydown",T),V.addEventListener("keyup",T),V.addEventListener("mouseleave",Q),window.addEventListener("blur",Q),()=>{V.removeEventListener("mousemove",N),V.removeEventListener("keydown",T),V.removeEventListener("keyup",T),V.removeEventListener("mouseleave",Q),window.removeEventListener("blur",Q)}},[ne,oe,je,fl,ut,Sl,he]);const ma=h.useCallback((x,S)=>{var N,T,Q,H;if(je&&!Qo){(N=Ja.current)==null||N.shake();return}if(oe&&!hl){const Y=(T=Ue.current)==null?void 0:T.querySelector("[data-annotation-popup]:not([data-annotation-card]) textarea");if(Y!=null&&Y.value.trim()){(Q=si.current)==null||Q.shake();return}ca(!0)}if(vn.current=S??null,ho.current=(S==null?void 0:S.matches(":focus-visible"))??!1,yn(!1),ys(!1),Gt(x),pn(null),At(null),ht([]),(H=x.elementBoundingBoxes)!=null&&H.length){const Y=[];for(const ce of x.elementBoundingBoxes){const me=ce.x+ce.width/2,ge=ce.y+ce.height/2-window.scrollY,Te=Hu(me,ge,ce);Te&&Y.push(Te)}dl(Y),Hn(null)}else if(x.boundingBox){const Y=x.boundingBox,ce=Y.x+Y.width/2,me=x.isFixed?Y.y+Y.height/2:Y.y+Y.height/2-window.scrollY,ge=Hu(ce,me,Y);if(ge){const Te=Un(ge),qe=Te.width/Y.width,Pe=Te.height/Y.height;qe<.5||Pe<.5?Hn(null):Hn(ge)}else Hn(null);dl([])}else Hn(null),dl([])},[oe,hl,je,Qo]);h.useEffect(()=>{if(!ne||fl||ut)return;const x=S=>{var st,St,$t;if(fa.current){fa.current=!1,S.preventDefault(),S.stopPropagation();return}const N=S.composedPath()[0]||S.target;if(ol(N,"[data-feedback-toolbar]")||ol(N,"[data-annotation-popup]")||ol(N,"[data-annotation-marker]"))return;if(Js(S)&&!oe&&!je){S.preventDefault(),S.stopPropagation(),hs.current=S.shiftKey;const Rn=B0(S.clientX,S.clientY);if(!Rn)return;const He=Un(Rn),{name:et,path:Ct,reactComponents:Wt}=Wu(Rn,Sl,he),il=Wn.findIndex(fn=>fn.element===Rn);il>=0?Vl(fn=>fn.filter((No,Zo)=>Zo!==il)):Vl(fn=>[...fn,{element:Rn,rect:He,name:et,path:Ct,reactComponents:Wt??void 0}]);return}const T=ol(N,"button, a, input, select, textarea, [role='button'], [onclick]");if(wt.blockInteractions&&(S.preventDefault(),S.stopPropagation()),oe&&!hl){if(T&&!wt.blockInteractions)return;S.preventDefault(),(st=si.current)==null||st.shake();return}if(je&&!Qo){if(T&&!wt.blockInteractions)return;S.preventDefault(),(St=Ja.current)==null||St.shake();return}S.preventDefault();const Q=ed(S.clientX,S.clientY);if(!Q)return;const{name:H,path:Y,reactComponents:ce}=Wu(Q,Sl,he),me=Un(Q),ge=S.clientX/window.innerWidth*100,Te=N_(Q),qe=Te?S.clientY:S.clientY+window.scrollY,Pe=($t=Q.ownerDocument.defaultView)==null?void 0:$t.getSelection();let dt;Pe&&Pe.toString().trim().length>0&&(dt=Pe.toString().trim().slice(0,500));const $e=Bu(Q),Ve=$u(Q);ca(!1),O({id:Date.now().toString(),x:ge,y:qe,clientY:S.clientY,element:H,elementPath:Y,selectedText:dt,boundingBox:{x:me.left,y:Te?me.top:me.top+window.scrollY,width:me.width,height:me.height},nearbyText:nc(Q),cssClasses:lc(Q),isFixed:Te,fullPath:cc(Q),accessibility:Uu(Q),computedStyles:Ve,computedStylesObj:$e,nearbyElements:Lu(Q),reactComponents:ce??void 0,sourceFile:Ku(Q),attributes:ic(Q,he),frame:F1(Q,S.clientX,S.clientY),targetElement:Q}),Fe(null)};return V.addEventListener("click",x,!0),()=>V.removeEventListener("click",x,!0)},[ne,fl,ut,oe,hl,je,Qo,wt.blockInteractions,Sl,he,Wn]),h.useEffect(()=>{if(!ne)return;const x=N=>{const T=(N.key==="Meta"||N.key==="Control")&&!Js(N),Q=N.key==="Shift"&&hs.current;(T||Q)&&!jl.current&&Wn.length>0&&bs()},S=()=>{var N;hs.current=!1,Vl([]),Fe(null),Go.current=null,jl.current=null,ua(!1),(N=Vo.current)==null||N.replaceChildren()};return V.addEventListener("keyup",x),window.addEventListener("blur",S),()=>{V.removeEventListener("keyup",x),window.removeEventListener("blur",S)}},[ne,Wn,bs]),h.useEffect(()=>{if(!ne||oe||fl||ut)return;const x=S=>{if(S.button!==0)return;fa.current=!1;const N=S.composedPath()[0]||S.target;if(ol(N,"[data-feedback-toolbar]")||ol(N,"[data-annotation-marker]")||ol(N,"[data-annotation-popup]"))return;const T=new Set(["P","SPAN","H1","H2","H3","H4","H5","H6","LI","TD","TH","LABEL","BLOCKQUOTE","FIGCAPTION","CAPTION","LEGEND","DT","DD","PRE","CODE","EM","STRONG","B","I","U","S","A","TIME","ADDRESS","CITE","Q","ABBR","DFN","MARK","SMALL","SUB","SUP"]);!Js(S)&&(T.has(N.tagName)||N.isContentEditable)||(S.preventDefault(),Go.current={x:S.clientX,y:S.clientY})};return V.addEventListener("mousedown",x),()=>V.removeEventListener("mousedown",x)},[ne,oe,fl,ut]),h.useEffect(()=>{if(!ne||oe)return;const x=S=>{if(!Go.current)return;const N=S.clientX-Go.current.x,T=S.clientY-Go.current.y,Q=N*N+T*T,H=oi*oi;if(!so&&Q>=H&&(jl.current=Go.current,ua(!0),S.preventDefault()),(so||Q>=H)&&jl.current){if(da.current){const He=Math.min(jl.current.x,S.clientX),et=Math.min(jl.current.y,S.clientY),Ct=Math.abs(S.clientX-jl.current.x),Wt=Math.abs(S.clientY-jl.current.y);da.current.style.transform=`translate(${He}px, ${et}px)`,da.current.style.width=`${Ct}px`,da.current.style.height=`${Wt}px`}const Y=Date.now();if(Y-bc.current<id)return;bc.current=Y;const ce=jl.current.x,me=jl.current.y,ge=Math.min(ce,S.clientX),Te=Math.min(me,S.clientY),qe=Math.max(ce,S.clientX),Pe=Math.max(me,S.clientY),dt=(ge+qe)/2,$e=(Te+Pe)/2,Ve=new Set,st=[[ge,Te],[qe,Te],[ge,Pe],[qe,Pe],[dt,$e],[dt,Te],[dt,Pe],[ge,$e],[qe,$e]];for(const[He,et]of st){const Ct=document.elementsFromPoint(He,et);for(const Wt of Ct)Wt instanceof HTMLElement&&Ve.add(Wt)}const St=V.querySelectorAll("button, a, input, img, p, h1, h2, h3, h4, h5, h6, li, label, td, th, div, span, section, article, aside, nav");for(const He of St)if(He instanceof HTMLElement){const et=Un(He),Ct=et.left+et.width/2,Wt=et.top+et.height/2,il=Ct>=ge&&Ct<=qe&&Wt>=Te&&Wt<=Pe,fn=Math.min(et.right,qe)-Math.max(et.left,ge),No=Math.min(et.bottom,Pe)-Math.max(et.top,Te),Zo=fn>0&&No>0?fn*No:0,zn=et.width*et.height,ks=zn>0?Zo/zn:0;(il||ks>.5)&&Ve.add(He)}const $t=[],Rn=new Set(["BUTTON","A","INPUT","IMG","P","H1","H2","H3","H4","H5","H6","LI","LABEL","TD","TH","SECTION","ARTICLE","ASIDE","NAV"]);for(const He of Ve){if(ol(He,"[data-feedback-toolbar]")||ol(He,"[data-annotation-marker]"))continue;const et=Un(He);if(!(et.width>window.innerWidth*.8&&et.height>window.innerHeight*.5)&&!(et.width<10||et.height<10)&&et.left<qe&&et.right>ge&&et.top<Pe&&et.bottom>Te){const Ct=He.tagName;let Wt=Rn.has(Ct);if(!Wt&&(Ct==="DIV"||Ct==="SPAN")){const il=He.textContent&&He.textContent.trim().length>0,fn=He.onclick!==null||He.getAttribute("role")==="button"||He.getAttribute("role")==="link"||He.classList.contains("clickable")||He.hasAttribute("data-clickable");(il||fn)&&!He.querySelector("p, h1, h2, h3, h4, h5, h6, button, a")&&(Wt=!0)}if(Wt){let il=!1;for(const fn of $t)if(fn.left<=et.left&&fn.right>=et.right&&fn.top<=et.top&&fn.bottom>=et.bottom){il=!0;break}il||$t.push(et)}}}if(Vo.current){const He=Vo.current;for(;He.children.length>$t.length;)He.removeChild(He.lastChild);$t.forEach((et,Ct)=>{let Wt=He.children[Ct];Wt||(Wt=document.createElement("div"),Wt.className=le.selectedElementHighlight,He.appendChild(Wt)),Wt.style.transform=`translate(${et.left}px, ${et.top}px)`,Wt.style.width=`${et.width}px`,Wt.style.height=`${et.height}px`})}}};return V.addEventListener("mousemove",x,{passive:!0}),()=>V.removeEventListener("mousemove",x)},[ne,oe,so,oi]),h.useEffect(()=>{if(!ne)return;const x=S=>{const N=so,T=jl.current;if(so&&T){fa.current=!0;const Q=Math.min(T.x,S.clientX),H=Math.min(T.y,S.clientY),Y=Math.max(T.x,S.clientX),ce=Math.max(T.y,S.clientY),me=[];V.querySelectorAll("button, a, input, img, p, h1, h2, h3, h4, h5, h6, li, label, td, th").forEach($e=>{if(!($e instanceof HTMLElement)||ol($e,"[data-feedback-toolbar]")||ol($e,"[data-annotation-marker]"))return;const Ve=Un($e);Ve.width>window.innerWidth*.8&&Ve.height>window.innerHeight*.5||Ve.width<10||Ve.height<10||Ve.left<Y&&Ve.right>Q&&Ve.top<ce&&Ve.bottom>H&&me.push({element:$e,rect:Ve})});const Te=me.filter(({element:$e})=>!me.some(({element:Ve})=>Ve!==$e&&$e.contains(Ve))),qe=S.clientX/window.innerWidth*100,Pe=S.clientY+window.scrollY,dt=(Js(S)||Wn.length>0)&&!oe&&!je;if(Te.length>0)if(dt){const $e=[...Wn];for(const{element:Ve,rect:st}of Te){if($e.some(He=>He.element===Ve))continue;const{name:St,path:$t,reactComponents:Rn}=Wu(Ve,Sl,he);$e.push({element:Ve,rect:st,name:St,path:$t,reactComponents:Rn??void 0})}hs.current=S.shiftKey,Js(S)?Vl($e):bs($e)}else{const $e=Te.reduce((He,{rect:et})=>({left:Math.min(He.left,et.left),top:Math.min(He.top,et.top),right:Math.max(He.right,et.right),bottom:Math.max(He.bottom,et.bottom)}),{left:1/0,top:1/0,right:-1/0,bottom:-1/0}),Ve=Te.slice(0,5).map(({element:He})=>qr(He).name).join(", "),st=Te.length>5?` +${Te.length-5} more`:"",St=Te[0].element,$t=Bu(St),Rn=$u(St);O({id:Date.now().toString(),x:qe,y:Pe,clientY:S.clientY,element:`${Te.length} elements: ${Ve}${st}`,elementPath:"multi-select",boundingBox:{x:$e.left,y:$e.top+window.scrollY,width:$e.right-$e.left,height:$e.bottom-$e.top},isMultiSelect:!0,fullPath:cc(St),accessibility:Uu(St),computedStyles:Rn,computedStylesObj:$t,nearbyElements:Lu(St),cssClasses:lc(St),nearbyText:nc(St),sourceFile:Ku(St),attributes:ic(St,he)})}else if(dt&&!Js(S))bs();else if(!dt){const $e=Math.abs(Y-Q),Ve=Math.abs(ce-H);$e>20&&Ve>20&&O({id:Date.now().toString(),x:qe,y:Pe,clientY:S.clientY,element:"Area selection",elementPath:`region at (${Math.round(Q)}, ${Math.round(H)})`,boundingBox:{x:Q,y:H+window.scrollY,width:$e,height:Ve},isMultiSelect:!0})}Fe(null)}else N&&(fa.current=!0);Go.current=null,jl.current=null,ua(!1),Vo.current&&(Vo.current.innerHTML="")};return V.addEventListener("mouseup",x),()=>V.removeEventListener("mouseup",x)},[ne,so,oe,je,Sl,he,Wn,bs]);const al=h.useCallback(async(x,S,N)=>{const T=wt.webhookUrl||W;if(!T||!wt.webhooksEnabled&&!N)return!1;try{return(await fetch(T,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({event:x,timestamp:Date.now(),url:typeof window<"u"?window.location.href:void 0,...S})})).ok}catch(Q){return console.warn("[Agentation] Webhook failed:",Q),!1}},[W,wt.webhookUrl,wt.webhooksEnabled]),ha=h.useCallback(x=>{var N;if(!oe||oe.isSubmitted)return;const S={id:oe.id,x:oe.x,y:oe.y,comment:x,element:oe.element,elementPath:oe.elementPath,timestamp:Date.now(),selectedText:oe.selectedText,boundingBox:oe.boundingBox,nearbyText:oe.nearbyText,cssClasses:oe.cssClasses,isMultiSelect:oe.isMultiSelect,isFixed:oe.isFixed,fullPath:oe.fullPath,accessibility:oe.accessibility,computedStyles:oe.computedStyles,nearbyElements:oe.nearbyElements,reactComponents:oe.reactComponents,sourceFile:oe.sourceFile,attributes:oe.attributes,frame:oe.frame,elementBoundingBoxes:oe.elementBoundingBoxes,...C&&dn?{sessionId:dn,url:typeof window<"u"?window.location.href:void 0,status:"pending"}:{}};Xt(T=>[...T,S]),O({...oe,isSubmitted:!0}),Ol.current=S.id,z==null||z(S),al("annotation.add",{annotation:S}),ca(!0),(N=window.getSelection())==null||N.removeAllRanges(),C&&dn&&G(async()=>{const T=await Se(C,dn,S);if(u){const Q=Zs(s);x_(s,Q.map(H=>H.id===S.id?{...H,id:T.id}:H),dn)}!ue.current||fe.current.has(S.id)||T.id!==S.id&&(Ka.current.set(T.id,S.id),Ol.current===S.id&&(Ol.current=T.id),Xt(Q=>Q.map(H=>H.id===S.id?{...H,id:T.id}:H)),Xo.current.delete(S.id)&&Xo.current.add(T.id))}).catch(T=>{console.warn("[Agentation] Failed to sync annotation:",T)})},[oe,z,al,C,dn,G,s,u]),ri=h.useCallback(()=>{ca(!0)},[]),vs=h.useCallback(()=>{O(null),ca(!1)},[]),sr=h.useCallback(x=>{if(fe.current.has(x))return;fe.current.add(x);const S=Ne.find(N=>N.id===x);(je==null?void 0:je.id)===x&&(yn(!1),ys(!0)),Za(N=>new Set(N).add(x)),S&&(E==null||E(S),al("annotation.delete",{annotation:S})),C&&G(()=>Gu(C,we.current.get(x)??x)).catch(N=>{console.warn("[Agentation] Failed to delete annotation from server:",N)})},[Ne,je,E,al,C,G]),ga=h.useCallback(x=>{var S;if(!x){pn(null),At(null),ht([]);return}if(pn(x.id),(S=x.elementBoundingBoxes)!=null&&S.length){const N=[];for(const T of x.elementBoundingBoxes){const Q=T.x+T.width/2,H=T.y+T.height/2-window.scrollY,Y=Hu(Q,H,T);Y&&N.push(Y)}ht(N),At(null)}else if(x.boundingBox){const N=x.boundingBox,T=N.x+N.width/2,Q=x.isFixed?N.y+N.height/2:N.y+N.height/2-window.scrollY,H=Hu(T,Q,N);if(H){const Y=Un(H),ce=Y.width/N.width,me=Y.height/N.height;ce<.5||me<.5?At(null):At(H)}else At(null);ht([])}else At(null),ht([])},[]),ii=h.useCallback(x=>{var N;if(!je)return;const S={...je,comment:x};Gt(S),Xt(T=>T.map(Q=>Q.id===je.id?S:Q)),$==null||$(S),al("annotation.update",{annotation:S}),C&&G(()=>w_(C,we.current.get(je.id)??je.id,{comment:x})).catch(T=>{console.warn("[Agentation] Failed to update annotation on server:",T)}),yn(ho.current||!!((N=vn.current)!=null&&N.matches(":hover"))),ys(!0)},[je,$,al,C,G]),ci=h.useCallback(()=>{var x;yn(ho.current||!!((x=vn.current)!=null&&x.matches(":hover"))),ys(!0)},[]),kc=h.useCallback(()=>{Ql&&je&&!oe&&pn(je.id),Gt(null),Hn(null),dl([]),ys(!1)},[Ql,je,oe]),pa=h.useCallback((x,S)=>{if(!x.length&&!S)return;ot(!0);const N={placements:[..._s.current.placements,...x],rearrange:S??_s.current.rearrange};_s.current=N,Jr(N.placements),qa(N.rearrange),clearTimeout(Ze.current),Ze.current=Ht(()=>{eo(T=>T.filter(Q=>!N.placements.includes(Q))),Rl(T=>T===N.rearrange?null:T),_s.current={placements:[],rearrange:null},Jr([]),qa(null),Ze.current=void 0,bt()},200)},[bt]),ko=h.useCallback(()=>{var ce;if(!ue.current)return;const x=new Map(jo.current.map(me=>[me.id,me])),S=[];for(const me of Ne){const ge=x.get(we.current.get(me.id)??me.id)??x.get(me.id);ge&&ge.comment===me.comment&&!fe.current.has(ge.id)&&!S.includes(ge)&&S.push(ge)}const N=S.length,T=Pr.current,Q=ft.filter(me=>T.designPlacements.includes(me)&&!_s.current.placements.includes(me)),H=be===T.rearrangeState&&be!==_s.current.rearrange?be:null,Y=_l.filter(me=>ei.current.includes(me));if(!(N===0&&Y.length===0&&Q.length===0&&!H)){for(const me of S)fe.current.add(me.id),ct.current.add(me.id),Ke.current.add(me.id);if(Za(me=>new Set([...me,...S.map(ge=>ge.id)])),B==null||B(S),al("annotations.clear",{annotations:S}),C&&Promise.all(S.map(me=>G(()=>Gu(C,we.current.get(me.id)??me.id)).catch(ge=>{console.warn("[Agentation] Failed to delete annotation from server:",ge)}))),ot(!0),rd(me=>me.filter(ge=>!Y.includes(ge))),Y.length>0&&Y.length===ei.current.length){const me=oa.current;(ce=me==null?void 0:me.getContext("2d"))==null||ce.clearRect(0,0,me.width,me.height)}pa(Q,H),Yt===T.blankCanvas&&qn===T.wireframePurpose&&ft===T.designPlacements&&be===T.rearrangeState&&(Yt&&ea(!1),qn&&ta(""),yt.current={rearrange:null,placements:[]},Qu(s)),bt()}},[s,Ne,_l,ft,be,Yt,qn,B,al,C,G,bt,pa]),ui=h.useCallback(async()=>{const x=ae.start(),S=typeof window<"u"?window.location.pathname+window.location.search+window.location.hash:s,N=ut&&Yt;let T;if(N){if(ft.length===0&&!be&&!qn)return;T=f?td(S,f):""}else{if(T=op(Ne,S,wt.outputDetail,{appName:f}),!T&&_l.length===0&&ft.length===0&&!be)return;T||(T=td(S,f))}if(!N&&_l.length>0){const H=new Set;for(const ge of Ne)ge.drawingIndex!=null&&H.add(ge.drawingIndex);const Y=oa.current;Y&&(Y.style.visibility="hidden");const ce=[],me=window.scrollY;for(let ge=0;ge<_l.length;ge++){if(H.has(ge))continue;const Te=_l[ge];if(Te.points.length<2)continue;const qe=Te.fixed?Te.points:Te.points.map(on=>({x:on.x,y:on.y-me}));let Pe=1/0,dt=1/0,$e=-1/0,Ve=-1/0;for(const on of qe)Pe=Math.min(Pe,on.x),dt=Math.min(dt,on.y),$e=Math.max($e,on.x),Ve=Math.max(Ve,on.y);const st=$e-Pe,St=Ve-dt,$t=Math.hypot(st,St),Rn=qe[0],He=qe[qe.length-1],et=Math.hypot(He.x-Rn.x,He.y-Rn.y);let Ct;const Wt=et<$t*.35,il=st/Math.max(St,1);if(Wt&&$t>20){const on=Math.max(st,St)*.15;let Ll=0;for(const Bl of qe){const kl=Bl.x-Pe<on,Cl=$e-Bl.x<on,Mo=Bl.y-dt<on,Wl=Ve-Bl.y<on;(kl||Cl)&&(Mo||Wl)&&Ll++}Ct=Ll>qe.length*.15?"box":"circle"}else il>3&&St<40?Ct="underline":et>$t*.5?Ct="arrow":Ct="drawing";const fn=Math.min(10,qe.length),No=Math.max(1,Math.floor(qe.length/fn)),Zo=new Set,zn=[],ks=[Rn];for(let on=No;on<qe.length-1;on+=No)ks.push(qe[on]);ks.push(He);for(const on of ks){const Ll=ed(on.x,on.y);if(!Ll||Zo.has(Ll)||ol(Ll,"[data-feedback-toolbar]"))continue;Zo.add(Ll);const{name:Bl}=qr(Ll);zn.includes(Bl)||zn.push(Bl)}const Zn=`${Math.round(Pe)},${Math.round(dt)} → ${Math.round($e)},${Math.round(Ve)}`;let Dl;(Ct==="circle"||Ct==="box")&&zn.length>0?Dl=`${Ct==="box"?"Boxed":"Circled"} **${zn[0]}**${zn.length>1?` (and ${zn.slice(1).join(", ")})`:""} (region: ${Zn})`:Ct==="underline"&&zn.length>0?Dl=`Underlined **${zn[0]}** (${Zn})`:Ct==="arrow"&&zn.length>=2?Dl=`Arrow from **${zn[0]}** to **${zn[zn.length-1]}** (${Math.round(Rn.x)},${Math.round(Rn.y)} → ${Math.round(He.x)},${Math.round(He.y)})`:zn.length>0?Dl=`${Ct==="arrow"?"Arrow":"Drawing"} near **${zn.join("**, **")}** (region: ${Zn})`:Dl=`Drawing at ${Zn}`,ce.push(Dl)}Y&&(Y.style.visibility=""),ce.length>0&&(T+=`
**Drawings:**
`,ce.forEach((ge,Te)=>{T+=`${Te+1}. ${ge}
`}))}if((ft.length>0||N&&qn)&&(T+=`
`+Z0(ft,{width:window.innerWidth,height:window.innerHeight},{blankCanvas:Yt,wireframePurpose:qn||void 0},wt.outputDetail)),be){const H=J0(be,wt.outputDetail,{width:window.innerWidth,height:window.innerHeight});H&&(T+=`
`+H)}if(T=sp(Ne,T,g),!T){P(!1);return}const Q=!te||await Z2(T);F==null||F(T),ae.isCurrent(x)&&(P(Q),Q&&(ae.schedule(x,()=>P(!1),2e3),wt.autoClearAfterCopy&&ae.schedule(x,ko,500)))},[Ne,_l,ft,be,Yt,ut,Tn,qn,s,wt.outputDetail,Sl,he,wt.autoClearAfterCopy,ko,ae,te,g,f,F]),di=fp(wt.webhookUrl)||fp(W||""),Fl=Me!=null||di&&!wt.webhooksEnabled,ar=ne?Fl?337:297:44,fi=h.useCallback(async()=>{const x=Be.start(),S=typeof window<"u"?window.location.href:s,N=typeof window<"u"?window.location.pathname+window.location.search+window.location.hash:s;let T=op(Ne,N,wt.outputDetail,{appName:f});if(!T&&ft.length===0&&!be)return;if(T||(T=td(N,f)),ft.length>0&&(T+=`
`+Z0(ft,{width:window.innerWidth,height:window.innerHeight},{blankCanvas:Yt,wireframePurpose:qn||void 0},wt.outputDetail)),be){const ce=J0(be,wt.outputDetail,{width:window.innerWidth,height:window.innerHeight});ce&&(T+=`
`+ce)}_e("sending");let Q=!0;try{await(Me==null?void 0:Me(T,Ne))}catch(ce){console.warn("[Agentation] Submit callback failed:",ce),Q=!1}if(!Be.isCurrent(x))return;const H=di?await al("submit",{output:T,annotations:Ne,url:S},!0):!0,Y=Q&&H&&Fl;Be.isCurrent(x)&&(_e(Y?"sent":"failed"),Be.schedule(x,()=>_e("idle"),2500),Y&&wt.autoClearAfterCopy&&Be.schedule(x,ko,500))},[Me,f,al,Ne,ft,be,Yt,Tn,s,wt.outputDetail,Sl,he,wt.autoClearAfterCopy,ko,di,Fl,Be]);h.useEffect(()=>{const S=(H=!1)=>{var Y;(Y=ia.current)!=null&&Y.dragging&&(ps.current=H,yc(!1)),ia.current=null},N=H=>{const Y=ia.current;if(!Y)return;if((H.buttons&1)===0){S();return}const ce=H.clientX-Y.x,me=H.clientY-Y.y,ge=Math.sqrt(ce*ce+me*me);if(!Y.dragging&&ge>10&&(Y.dragging=!0,yc(!0)),Y.dragging){let Te=Y.toolbarX+ce,qe=Y.toolbarY+me;const Pe=20,dt=337,$e=44,st=dt-ar,St=Pe-st,$t=window.innerWidth-Pe-dt;Te=Math.max(St,Math.min($t,Te)),qe=Math.max(Pe,Math.min(window.innerHeight-$e-Pe,qe)),Fa({x:Te,y:qe})}},T=()=>S(!0),Q=()=>S();return V.addEventListener("mousemove",N),V.addEventListener("mouseup",T,!0),window.addEventListener("blur",Q),()=>{V.removeEventListener("mousemove",N),V.removeEventListener("mouseup",T,!0),window.removeEventListener("blur",Q)}},[ar]);const gl=h.useCallback(x=>{if(ps.current=!1,ia.current=null,x.button!==0||x.target.closest("button")&&(x.target.closest("button")!==un.current||ne)||x.target.closest("[data-agentation-settings-panel]"))return;const S=x.currentTarget.parentElement;if(!S)return;const N=Un(S);ia.current={x:x.clientX,y:x.clientY,toolbarX:N.left,toolbarY:N.top,dragging:!1}},[ne]);h.useLayoutEffect(()=>{if(!ln)return;const x=()=>{let Q=ln.x,H=ln.y;const me=20-(337-ar),ge=window.innerWidth-20-337;Q=Math.max(me,Math.min(ge,Q)),H=Math.max(20,Math.min(window.innerHeight-44-20,H)),(Q!==ln.x||H!==ln.y)&&Fa({x:Q,y:H})};return x(),window.addEventListener("resize",x),()=>window.removeEventListener("resize",x)},[ln,ar]),h.useEffect(()=>{if(!d)return;const x=N=>{var H,Y;if(N.defaultPrevented||N.isComposing||N.altKey)return;const T=N.composedPath()[0]||N.target,Q=T.tagName==="INPUT"||T.tagName==="TEXTAREA"||T.tagName==="SELECT"||T.isContentEditable;if(N.key==="Escape"){if(j&&!oe&&!je&&(ne||Yn||ut||fl||Wn.length)&&(N.preventDefault(),N.stopPropagation()),Yn){N.preventDefault(),go(!1),(H=Cn.current)==null||H.focus();return}if(ut){is?po(null):xs();return}if(fl){Qa(!1);return}if(Wn.length>0){Vl([]);return}oe||je||ne&&(Sn(),ao())}if((N.metaKey||N.ctrlKey)&&N.shiftKey&&(N.key==="f"||N.key==="F")){N.preventDefault(),Sn(),ne?ao():((Y=un.current)==null||Y.blur(),bn.current=!0,Ce(!0));return}!ne||Q||N.metaKey||N.ctrlKey||N.repeat||((N.key==="p"||N.key==="P")&&(N.preventDefault(),Sn(),ai()),(N.key==="l"||N.key==="L")&&(N.preventDefault(),Sn(),fl&&Qa(!1),Yn&&go(!1),oe&&ri(),ut?xs():lr()),(N.key==="h"||N.key==="H")&&Ne.length>0&&(N.preventDefault(),Sn(),We(ce=>!ce)),(N.key==="c"||N.key==="C")&&(Ne.length>0||ft.length>0||be)&&(N.preventDefault(),Sn(),ui()),(N.key==="x"||N.key==="X")&&(Ne.length>0||ft.length>0||be)&&(N.preventDefault(),Sn(),ko(),ft.length>0&&eo([]),be&&Rl(null)),(N.key==="s"||N.key==="S")&&Ne.length>0&&Fl&&ke==="idle"&&(N.preventDefault(),Sn(),fi()))},S=!!j;return V.addEventListener("keydown",x,S),()=>V.removeEventListener("keydown",x,S)},[d,j,je,ne,fl,ut,is,ft,be,oe,Ne.length,Fl,ke,fi,ai,ui,ko,Wn,Yn,ao,lr,xs]);const ro=Ne.length>0,ws=W1(),ya=Ne.filter(x=>x.kind!=="placement"&&x.kind!=="rearrange"),_i=ya.flatMap((x,S)=>{const N=ws(x);return N?[{annotation:N,index:S}]:[]}),mi=oe&&!oe.isSubmitted?ws({...oe,comment:"",timestamp:0}):null,hi=[...Ge?_i.map(x=>({...x,pending:!1})):[],...mi?[{annotation:mi,index:ya.length,pending:!0}]:[]];h.useEffect(()=>{const x=new Set(Ge&&!gn?_i.map(({annotation:S})=>S.id):[]);Ol.current&&!x.has(Ol.current)&&(Ol.current=null);for(const S of So)x.has(S)||tr(S)}),h.useEffect(()=>{je&&So.has(je.id)&&(yn(!1),ys(!0))},[je,So]);const rr=h.useCallback(x=>{!lt&&x.id!==Ol.current&&ga(x)},[lt,ga]),Cc=h.useCallback(x=>{Bt===x&&ga(null)},[Bt,ga]),Fo=h.useCallback((x,S)=>{var N;if(je&&!Qo){(N=Ja.current)==null||N.shake();return}hl&&vs(),wt.markerClickBehavior==="delete"?sr(x.id):ma(x,S)},[wt.markerClickBehavior,sr,ma,hl,vs,je,Qo]),rl=je??(Pa&&!oe&&!at?Ne.find(x=>x.id===Bt&&!So.has(x.id)):null),Ss=vl?"Resume animations":"Pause animations",ir=ut?"Exit layout mode":"Layout mode",gi=xe?"Hide markers":"Show markers",Wo=g!=="markdown"&&!sp(Ne,"",g),cr=typeof g=="object"?`Copy ${g.attribute}`:g==="source"?"Copy source paths":g==="classes"?"Copy classes":ut&&Yt?"Copy layout":"Copy feedback",Ko=ne?0:-1;return!Ot||gn?null:o.jsxs(_v,{host:"agentation-toolbar",className:Ee,children:[o.jsxs("style",{"data-agentation-styles":"toolbar",children:[fw,_w]}),o.jsxs("div",{ref:Ue,className:le.positionContext,style:{display:"contents"},"data-agentation-theme":ml?"dark":"light","data-agentation-accent":wt.annotationColorId,"data-agentation-root":"",children:[o.jsx("div",{className:le.toolbar,"data-feedback-toolbar":!0,"data-agentation-toolbar":!0,"data-dragging":Wa||void 0,style:ln?{left:ln.x,top:ln.y,right:"auto",bottom:"auto"}:void 0,children:o.jsxs("div",{className:`${le.toolbarContainer} ${ne?le.expanded:le.collapsed} ${aa?le.entrance:""} ${mn?le.hiding:""} ${Fl?le.serverConnected:""}`,onMouseDown:gl,children:[o.jsxs("div",{className:`${le.controlsContent} ${ne?le.visible:le.hidden} ${ln&&ln.y<100?le.tooltipBelow:""} ${Kr||Yn?le.tooltipsHidden:""} ${sa?le.tooltipsInSession:""}`,ref:x=>{ul.current=x,x==null||x.toggleAttribute("inert",!ne)},role:"group","aria-label":"Feedback controls","aria-hidden":!ne,onMouseEnter:Ga,onMouseLeave:gs,children:[o.jsxs("div",{className:`${le.buttonWrapper} ${ln&&ln.x<120?le.buttonWrapperAlignLeft:""}`,children:[o.jsx("button",{className:le.controlButton,onClick:x=>{x.stopPropagation(),Sn(),ai()},"data-active":vl,"aria-label":Ss,"aria-pressed":vl,tabIndex:Ko,children:o.jsx(xv,{size:24,isPaused:vl})}),o.jsxs("span",{className:le.buttonTooltip,children:[Ss,d&&o.jsx("span",{className:le.shortcut,children:"P"})]})]}),o.jsxs("div",{className:le.buttonWrapper,children:[o.jsx("button",{className:`${le.controlButton} ${ml?"":le.light}`,onClick:x=>{x.stopPropagation(),Sn(),fl&&Qa(!1),Yn&&go(!1),oe&&ri(),ut?xs():lr()},"data-active":ut,"aria-label":ir,"aria-pressed":ut,tabIndex:Ko,style:ut&&Yt?{color:"#f97316",background:"rgba(249, 115, 22, 0.25)"}:void 0,children:o.jsx(Nv,{size:21})}),o.jsxs("span",{className:le.buttonTooltip,children:[ir,d&&o.jsx("span",{className:le.shortcut,children:"L"})]})]}),o.jsxs("div",{className:le.buttonWrapper,children:[o.jsx("button",{className:le.controlButton,onClick:x=>{x.stopPropagation(),Sn(),We(!xe)},disabled:!ro||ut,"aria-label":gi,tabIndex:Ko,children:o.jsx(yv,{size:24,isOpen:xe})}),o.jsxs("span",{className:le.buttonTooltip,children:[gi,d&&o.jsx("span",{className:le.shortcut,children:"H"})]})]}),o.jsxs("div",{className:le.buttonWrapper,children:[o.jsx("button",{className:`${le.controlButton} ${I?le.statusShowing:""}`,onClick:x=>{x.stopPropagation(),Sn(),ui()},disabled:Wo||(ut&&Yt?ft.length===0&&!((Co=be==null?void 0:be.sections)!=null&&Co.length):!ro&&_l.length===0&&ft.length===0&&!((pi=be==null?void 0:be.sections)!=null&&pi.length)),"data-active":I,"aria-label":cr,tabIndex:Ko,children:o.jsx(gv,{size:24,copied:I,tint:ut&&Yt&&(ft.length>0||(yi=be==null?void 0:be.sections)!=null&&yi.length)?"#f97316":void 0})}),o.jsxs("span",{className:le.buttonTooltip,children:[Wo?"No matching metadata":cr,d&&o.jsx("span",{className:le.shortcut,children:"C"})]})]}),o.jsxs("div",{className:`${le.buttonWrapper} ${le.sendButtonWrapper} ${ne&&Fl?le.sendButtonVisible:""}`,children:[o.jsxs("button",{className:`${le.controlButton} ${ke==="sent"||ke==="failed"?le.statusShowing:""}`,onClick:x=>{x.stopPropagation(),Sn(),fi()},disabled:!ro||!Fl||ke==="sending","data-no-hover":ke==="sent"||ke==="failed",tabIndex:ne&&Fl?0:-1,"aria-label":"Send Annotations","aria-hidden":!Fl,children:[o.jsx(pv,{size:24,state:ke}),ro&&ke==="idle"&&o.jsx("span",{className:le.buttonBadge,children:Ne.length})]}),o.jsxs("span",{className:le.buttonTooltip,children:["Send Annotations",d&&o.jsx("span",{className:le.shortcut,children:"S"})]})]}),o.jsxs("div",{className:le.buttonWrapper,children:[o.jsx("button",{className:le.controlButton,onClick:x=>{x.stopPropagation(),Sn(),ko()},disabled:!ro&&_l.length===0&&ft.length===0&&!((xi=be==null?void 0:be.sections)!=null&&xi.length),"data-danger":!0,"aria-label":"Clear all",tabIndex:Ko,children:o.jsx(vv,{size:24})}),o.jsxs("span",{className:le.buttonTooltip,children:["Clear all",d&&o.jsx("span",{className:le.shortcut,children:"X"})]})]}),o.jsxs("div",{className:le.buttonWrapper,children:[o.jsx("button",{ref:Cn,"aria-label":"Settings","aria-expanded":Yn,tabIndex:Ko,className:le.controlButton,onClick:x=>{x.stopPropagation(),Sn(),ut&&xs(),ye.current=!Yn&&x.detail===0,go(!Yn)},children:o.jsx(bv,{size:24})}),C&&oo!=="disconnected"&&o.jsx("span",{className:`${le.mcpIndicator} ${le[oo]} ${Yn?le.hidden:""}`,title:oo==="connected"?"MCP Connected":"MCP Connecting..."}),o.jsx("span",{className:le.buttonTooltip,children:"Settings"})]}),o.jsx("div",{className:le.divider}),o.jsx("div",{className:le.togglePlaceholder,"aria-hidden":"true"})]}),o.jsxs("div",{className:`${le.buttonWrapper} ${le.toggleWrapper} ${ln&&ln.y<100?le.tooltipBelow:""} ${!ne||Kr||Yn?le.tooltipsHidden:""} ${sa?le.tooltipsInSession:""} ${ln&&typeof window<"u"&&ln.x>window.innerWidth-120?le.buttonWrapperAlignRight:""}`,onMouseEnter:Ga,onMouseLeave:gs,children:[o.jsx("button",{ref:un,type:"button",className:`${le.toggleContent} ${ne?le.expandedToggle:""}`,"aria-label":ne?"Exit":"Start feedback mode","aria-expanded":ne,"aria-keyshortcuts":d?"Meta+Shift+F Control+Shift+F":void 0,title:ne?void 0:d?"Start feedback mode (⌘⇧F / Ctrl+Shift+F)":"Start feedback mode",onClick:x=>{if(ps.current){ps.current=!1,x.preventDefault();return}x.stopPropagation(),ne?(Sn(),ao()):(x.currentTarget.blur(),bn.current=x.detail===0,Ce(!0))},children:o.jsxs("span",{className:le.toggleIcon,children:[o.jsx(Rv,{active:ne}),ya.length>0&&o.jsx("span",{className:`${le.badge} ${ne?le.fadeOut:""} ${aa?le.entrance:""}`,children:ya.length})]})}),o.jsxs("span",{className:le.buttonTooltip,"aria-hidden":!ne,children:["Exit",d&&o.jsx("span",{className:le.shortcut,children:"Esc"})]})]}),o.jsx(J5,{visible:ut&&ne,activeType:is,onSelect:x=>{po(is===x?null:x)},isDarkMode:ml,sectionCount:(be==null?void 0:be.sections.length)??0,onDetectSections:()=>{const x=n2(),S=(be==null?void 0:be.sections)??[],N=new Set(S.map(Y=>Y.selector)),T=x.filter(Y=>!N.has(Y.selector)),Q=[...S,...T],H=[...(be==null?void 0:be.originalOrder)??[],...T.map(Y=>Y.id)];Rl({sections:Q,originalOrder:H,detectedAt:Date.now()})},placementCount:ft.length,onClearPlacements:()=>{pa(ft,be)},blankCanvas:Yt,onBlankCanvasChange:x=>{const S={sections:[],originalOrder:[],detectedAt:Date.now()};x?(Zr.current={rearrange:be,placements:ft},Rl(yt.current.rearrange||S),eo(yt.current.placements),po(null)):(yt.current={rearrange:be,placements:ft},Rl(Zr.current.rearrange||S),eo(Zr.current.placements)),ea(x)},wireframePurpose:qn,onWireframePurposeChange:ta,Tooltip:La,onDragStart:(x,S)=>{S.preventDefault();const N=Ae[x];let T=null,Q=!1;const H=S.clientX,Y=S.clientY,ce=S.target.closest("[data-feedback-toolbar]"),me=(ce==null?void 0:ce.getBoundingClientRect().top)??window.innerHeight,ge=qe=>{var Wt;const Pe=qe.clientX-H,dt=qe.clientY-Y;if(!Q&&(Math.abs(Pe)>4||Math.abs(dt)>4)&&(Q=!0,T=document.createElement("div"),T.className=`${ee.dragPreview}${Yt?` ${ee.dragPreviewWireframe}`:""}`,(Wt=Ue.current)==null||Wt.appendChild(T)),!T)return;const $e=Math.max(0,me-qe.clientY),Ve=Math.min(1,$e/180),st=1-Math.pow(1-Ve,2),St=28,$t=20,Rn=Math.min(140,N.width*.18),He=Math.min(90,N.height*.18),et=St+(Rn-St)*st,Ct=$t+(He-$t)*st;T.style.width=`${et}px`,T.style.height=`${Ct}px`,T.style.left=`${qe.clientX-et/2}px`,T.style.top=`${qe.clientY-Ct/2}px`,T.style.opacity=`${.5+.5*st}`,T.textContent=st>.25?x:""},Te=qe=>{if(window.removeEventListener("mousemove",ge),window.removeEventListener("mouseup",Te),T&&T.remove(),Q){const Pe=N.width,dt=N.height,$e=window.scrollY,Ve=Math.max(0,qe.clientX-Pe/2),st=Math.max(0,qe.clientY+$e-dt/2),St={id:`dp-${Date.now()}-${Math.random().toString(36).slice(2,7)}`,type:x,x:Ve,y:st,width:Pe,height:dt,scrollY:$e,timestamp:Date.now()};eo($t=>[...$t,St]),po(null),Xn.current=new Set,na($t=>$t+1)}};window.addEventListener("mousemove",ge),window.addEventListener("mouseup",Te)}}),o.jsx(cw,{settings:wt,onSettingsChange:pc,isDarkMode:ml,onToggleTheme:Kn,isDevMode:wo,connectionStatus:oo,endpoint:C,onExited:cd,isOpen:ne&&Yn,toolbarNearBottom:!!ln&&ln.y<230,settingsPage:Wr,onSettingsPageChange:$a,onHideToolbar:Sc})]})}),(ut||wn)&&o.jsx("div",{className:`${ee.blankCanvas} ${yo?ee.visible:""} ${Ho?ee.gridActive:""}`,style:{"--canvas-opacity":cs},"data-feedback-toolbar":!0}),ut&&Yt&&yo&&o.jsxs("div",{className:ee.wireframeNotice,"data-feedback-toolbar":!0,children:[o.jsxs("div",{className:ee.wireframeOpacityRow,children:[o.jsx("span",{className:ee.wireframeOpacityLabel,children:"Toggle Opacity"}),o.jsx("input",{type:"range",className:ee.wireframeOpacitySlider,min:0,max:1,step:.01,value:cs,onChange:x=>us(Number(x.target.value))})]}),o.jsxs("div",{className:ee.wireframeNoticeTitleRow,children:[o.jsx("span",{className:ee.wireframeNoticeTitle,children:"Wireframe Mode"}),o.jsx("span",{className:ee.wireframeNoticeDivider}),o.jsx("button",{className:ee.wireframeStartOver,onClick:()=>{pa(ft,be),yt.current={rearrange:null,placements:[]},ta(""),Qu(s)},children:"Start Over"})]}),"Drag components onto the canvas.",o.jsx("br",{}),"Copied output will only include the wireframed layout."]}),(ut||wn)&&o.jsx(V5,{placements:ft,onChange:eo,activeComponent:wn?null:is,onActiveComponentChange:po,isDarkMode:ml,exiting:wn,onInteractionChange:ds,passthrough:!is,extraSnapRects:be==null?void 0:be.sections.map(x=>x.currentRect),deselectSignal:gc,clearingPlacements:On,wireframe:Yt,onSelectionChange:(x,S)=>{Xn.current=x,S||(ms.current=new Set,to(N=>N+1))},onDragMove:(x,S)=>{var T;const N=ms.current;if(!(!N.size||!be)){if(!wl.current){wl.current=new Map;for(const Q of be.sections)N.has(Q.id)&&wl.current.set(Q.id,{x:Q.currentRect.x,y:Q.currentRect.y})}for(const Q of be.sections){if(!N.has(Q.id)||!wl.current.get(Q.id))continue;const Y=(T=Ue.current)==null?void 0:T.querySelector(`[data-rearrange-section="${Q.id}"]`);Y&&(Y.style.transform=`translate(${x}px, ${S}px)`)}}},onDragEnd:(x,S,N)=>{var H;const T=ms.current,Q=wl.current;if(wl.current=null,!(!T.size||!be||!Q)){for(const Y of T){const ce=(H=Ue.current)==null?void 0:H.querySelector(`[data-rearrange-section="${Y}"]`);ce&&(ce.style.transform="")}N&&Rl(Y=>Y&&{...Y,sections:Y.sections.map(ce=>{const me=Q.get(ce.id);return me?{...ce,currentRect:{...ce.currentRect,x:Math.max(0,me.x+x),y:Math.max(0,me.y+S)}}:ce})})}}}),(ut||wn)&&be&&o.jsx(s2,{rearrangeState:be,onChange:Rl,isDarkMode:ml,exiting:wn,blankCanvas:Yt,extraSnapRects:ft.map(x=>({x:x.x,y:x.y,width:x.width,height:x.height})),clearing:be===fs,deselectSignal:Ya,onSelectionChange:(x,S)=>{ms.current=x,S||(Xn.current=new Set,na(N=>N+1))},onDragMove:(x,S)=>{var T;const N=Xn.current;if(N.size){if(!wl.current){wl.current=new Map;for(const Q of ft)N.has(Q.id)&&wl.current.set(Q.id,{x:Q.x,y:Q.y})}for(const Q of N){const H=(T=Ue.current)==null?void 0:T.querySelector(`[data-design-placement="${Q}"]`);H&&(H.style.transform=`translate(${x}px, ${S}px)`)}}},onDragEnd:(x,S,N)=>{var H;const T=Xn.current,Q=wl.current;if(wl.current=null,!(!T.size||!Q)){for(const Y of T){const ce=(H=Ue.current)==null?void 0:H.querySelector(`[data-design-placement="${Y}"]`);ce&&(ce.style.transform="")}N&&eo(Y=>Y.map(ce=>{const me=Q.get(ce.id);return me?{...ce,x:Math.max(0,me.x+x),y:Math.max(0,me.y+S)}:ce}))}}}),o.jsx("canvas",{ref:oa,className:`${le.drawCanvas} ${fl?le.active:""}`,"aria-hidden":"true",style:{opacity:Pa?1:0,transition:"opacity 0.15s ease"},"data-feedback-toolbar":!0}),o.jsx("div",{className:le.markersLayer,"data-feedback-toolbar":!0,children:hi.filter(({annotation:x})=>!x.isFixed).map(({annotation:x,index:S,pending:N},T,Q)=>o.jsx(ip,{annotation:x,pending:N,globalIndex:S,layerIndex:T,layerSize:Q.length,isExiting:N?hl:lt,isClearing:ct.current.has(x.id),isAnimated:Xo.current.has(x.id),isNew:Ol.current===x.id,onEnterComplete:xc,isHovered:!lt&&Bt===x.id,isRemoving:So.has(x.id),onRemoveComplete:tr,isEditingAny:!!je,renumberFrom:Nn,markerClickBehavior:wt.markerClickBehavior,onHoverEnter:rr,onHoverLeave:Cc,onClick:Fo,onContextMenu:ma},Ka.current.get(x.id)??x.id))}),o.jsx("div",{className:le.fixedMarkersLayer,"data-feedback-toolbar":!0,children:hi.filter(({annotation:x})=>x.isFixed).map(({annotation:x,index:S,pending:N},T,Q)=>o.jsx(ip,{annotation:x,pending:N,globalIndex:S,layerIndex:T,layerSize:Q.length,isExiting:N?hl:lt,isClearing:ct.current.has(x.id),isAnimated:Xo.current.has(x.id),isNew:Ol.current===x.id,onEnterComplete:xc,isHovered:!lt&&Bt===x.id,isRemoving:So.has(x.id),onRemoveComplete:tr,isEditingAny:!!je,renumberFrom:Nn,markerClickBehavior:wt.markerClickBehavior,onHoverEnter:rr,onHoverLeave:Cc,onClick:Fo,onContextMenu:ma},Ka.current.get(x.id)??x.id))}),ne&&it&&!oe&&!je&&!hc&&!so&&o.jsx(uw,{x:Nt.x,y:Nt.y,elementName:it.elementName,reactComponents:it.reactComponents}),ne&&o.jsxs("div",{className:le.overlay,"data-feedback-toolbar":!0,style:oe||je?{zIndex:"inherit"}:void 0,children:[(it==null?void 0:it.rect)&&!oe&&!hc&&!so&&o.jsx("div",{className:`${le.hoverHighlight} ${le.enter}`,style:{left:it.rect.left,top:it.rect.top,width:it.rect.width,height:it.rect.height,borderColor:"color-mix(in srgb, var(--agentation-color-accent) 50%, transparent)",backgroundColor:"color-mix(in srgb, var(--agentation-color-accent) 4%, transparent)",...it.isPiercing?{borderStyle:"dashed"}:{}}}),Wn.filter(x=>x.element.isConnected).map((x,S)=>{const N=Un(x.element),T=Wn.length>1;return o.jsx("div",{className:T?le.multiSelectOutline:le.singleSelectOutline,style:{position:"fixed",left:N.left,top:N.top,width:N.width,height:N.height,...T?{}:{borderColor:"color-mix(in srgb, var(--agentation-color-accent) 60%, transparent)",backgroundColor:"color-mix(in srgb, var(--agentation-color-accent) 5%, transparent)"}}},S)}),Bt&&!oe&&(()=>{var Q;const x=Ne.find(H=>H.id===Bt);if(!(x!=null&&x.boundingBox))return null;if((Q=x.elementBoundingBoxes)!=null&&Q.length)return kt.length>0?kt.filter(H=>H.isConnected).map((H,Y)=>{const ce=Un(H);return o.jsx("div",{className:`${le.multiSelectOutline} ${le.enter}`,style:{left:ce.left,top:ce.top,width:ce.width,height:ce.height}},`hover-outline-live-${Y}`)}):x.elementBoundingBoxes.map((H,Y)=>o.jsx("div",{className:`${le.multiSelectOutline} ${le.enter}`,style:{left:H.x,top:H.y-sl,width:H.width,height:H.height}},`hover-outline-${Y}`));const S=pt&&pt.isConnected?Un(pt):null,N=S?{x:S.left,y:S.top,width:S.width,height:S.height}:{x:x.boundingBox.x,y:x.isFixed?x.boundingBox.y:x.boundingBox.y-sl,width:x.boundingBox.width,height:x.boundingBox.height},T=x.isMultiSelect;return o.jsx("div",{className:`${T?le.multiSelectOutline:le.singleSelectOutline} ${le.enter}`,style:{left:N.x,top:N.y,width:N.width,height:N.height,...T?{}:{borderColor:"color-mix(in srgb, var(--agentation-color-accent) 60%, transparent)",backgroundColor:"color-mix(in srgb, var(--agentation-color-accent) 5%, transparent)"}}})})(),oe&&o.jsxs(o.Fragment,{children:[(Nc=oe.multiSelectElements)!=null&&Nc.length?oe.multiSelectElements.filter(x=>x.isConnected).map((x,S)=>{const N=Un(x);return o.jsx("div",{className:`${le.multiSelectOutline} ${hl?le.exit:le.enter}`,style:{left:N.left,top:N.top,width:N.width,height:N.height}},`pending-multi-${S}`)}):oe.targetElement&&oe.targetElement.isConnected?(()=>{const x=Un(oe.targetElement);return o.jsx("div",{className:`${le.singleSelectOutline} ${hl?le.exit:le.enter}`,style:{left:x.left,top:x.top,width:x.width,height:x.height,borderColor:"color-mix(in srgb, var(--agentation-color-accent) 60%, transparent)",backgroundColor:"color-mix(in srgb, var(--agentation-color-accent) 5%, transparent)"}})})():oe.boundingBox&&o.jsx("div",{className:`${oe.isMultiSelect?le.multiSelectOutline:le.singleSelectOutline} ${hl?le.exit:le.enter}`,style:{left:oe.boundingBox.x,top:oe.boundingBox.y-sl,width:oe.boundingBox.width,height:oe.boundingBox.height,...oe.isMultiSelect?{}:{borderColor:"color-mix(in srgb, var(--agentation-color-accent) 60%, transparent)",backgroundColor:"color-mix(in srgb, var(--agentation-color-accent) 5%, transparent)"}}}),(()=>{var T,Q;const x=ws(oe)??oe,S=x.x,N=x.isFixed?x.y:x.y-sl;return o.jsx(o.Fragment,{children:o.jsx(K_,{ref:si,element:oe.element,selectedText:oe.selectedText,allowEmpty:typeof g=="object"&&!!((T=oe.attributes)!=null&&T[g.attribute]),onOpenSource:p&&oe.sourceFile?()=>p(oe.sourceFile):void 0,computedStyles:oe.computedStylesObj,placeholder:typeof g=="object"&&((Q=oe.attributes)!=null&&Q[g.attribute])?"Add a note (optional)":oe.element==="Area selection"?"What should change in this area?":oe.isMultiSelect?"Feedback for this group of elements...":"What should change?",onSubmit:ha,onExitComplete:vs,onCancel:ri,isExiting:hl,lightMode:!ml,accentColor:oe.isMultiSelect?"var(--agentation-color-green)":"var(--agentation-color-accent)",style:{left:Math.max(160,Math.min(window.innerWidth-160,S/100*window.innerWidth)),...N>window.innerHeight-290?{bottom:window.innerHeight-N+20}:{top:N+20}}},oe.id)})})()]}),je&&o.jsx(o.Fragment,{children:(Mc=je.elementBoundingBoxes)!=null&&Mc.length?Mn.length>0?Mn.filter(x=>x.isConnected).map((x,S)=>{const N=Un(x);return o.jsx("div",{className:`${le.multiSelectOutline} ${le.enter}`,style:{left:N.left,top:N.top,width:N.width,height:N.height}},`edit-multi-live-${S}`)}):je.elementBoundingBoxes.map((x,S)=>o.jsx("div",{className:`${le.multiSelectOutline} ${le.enter}`,style:{left:x.x,top:x.y-sl,width:x.width,height:x.height}},`edit-multi-${S}`)):(()=>{const x=Fn&&Fn.isConnected?Un(Fn):null,S=x?{x:x.left,y:x.top,width:x.width,height:x.height}:je.boundingBox?{x:je.boundingBox.x,y:je.isFixed?je.boundingBox.y:je.boundingBox.y-sl,width:je.boundingBox.width,height:je.boundingBox.height}:null;return S?o.jsx("div",{className:`${je.isMultiSelect?le.multiSelectOutline:le.singleSelectOutline} ${le.enter}`,style:{left:S.x,top:S.y,width:S.width,height:S.height,...je.isMultiSelect?{}:{borderColor:"color-mix(in srgb, var(--agentation-color-accent) 60%, transparent)",backgroundColor:"color-mix(in srgb, var(--agentation-color-accent) 5%, transparent)"}}}):null})()}),so&&o.jsxs(o.Fragment,{children:[o.jsx("div",{ref:da,className:le.dragSelection}),o.jsx("div",{ref:Vo,className:le.highlightsContainer})]})]}),o.jsx(nw,{ref:Ja,annotation:rl?ws(rl)??je:null,editing:!!je,exiting:Qo,restorePreview:Ql,scrollY:sl,lightMode:!ml,onExited:kc,editorProps:rl?{element:rl.element,selectedText:rl.selectedText,allowEmpty:typeof g=="object"&&!!((Ec=rl.attributes)!=null&&Ec[g.attribute]),onOpenSource:p&&rl.sourceFile?()=>p(rl.sourceFile):void 0,computedStyles:av(rl.computedStyles),placeholder:"Edit your feedback...",initialValue:rl.comment,submitLabel:"Save",onSubmit:ii,onCancel:ci,onDelete:()=>sr(rl.id),accentColor:rl.isMultiSelect?"var(--agentation-color-green)":"var(--agentation-color-accent)"}:void 0})]})]})}const _p=h.lazy(()=>_c(()=>import("./ProfilesView-DTwSR_35.js"),__vite__mapDeps([0,1,2,3,4,5,6,7,8])).then(s=>({default:s.ProfilesView}))),gw=h.lazy(()=>_c(()=>import("./SessionsView-BJNgU4jK.js"),__vite__mapDeps([9,1,2,10,6,8,7])).then(s=>({default:s.SessionsView}))),pw=h.lazy(()=>_c(()=>import("./SessionDetailView-CqdZp2XU.js"),__vite__mapDeps([11,1,2,7,8,6,12,5])).then(s=>({default:s.SessionDetailView}))),yw=h.lazy(()=>_c(()=>import("./McpView-vcf5UOVw.js"),__vite__mapDeps([13,1,2,3,6])).then(s=>({default:s.McpView}))),xw=h.lazy(()=>_c(()=>import("./NetworkView-4sBoSKvp.js"),__vite__mapDeps([14,1,2,3,7,6,5,4])).then(s=>({default:s.NetworkView})));function mp(){try{const s=window.location.pathname.replace(/\/+$/,"")||"/",r=new URLSearchParams(window.location.search),c=r.get("session")||r.get("id"),u=r.get("view"),f=r.get("profile"),d=r.get("tab")||"quota",m=r.get("scope")||"all",g=d==="telemetry"?"quota":d;if(s.startsWith("/sessions/")&&s.length>10){const j=s.slice(10).split("/");let k=decodeURIComponent(j[j.length-1]),b;return j.length>=2&&(b=decodeURIComponent(j[0]),k=decodeURIComponent(j.slice(1).join("/"))),{view:"session-detail",profile:f||"all",tab:g,mcpScope:m,sessionId:k,agentHint:b}}if(s==="/sessions")return{view:"sessions",profile:f||"all",tab:g,mcpScope:m,sessionId:null};if(s==="/mcp")return{view:"mcp",profile:f||"all",tab:g,mcpScope:m!=="all"?m:f&&f!=="all"?f:"all",sessionId:null};if(s==="/network")return{view:"network",profile:f||"all",tab:g,mcpScope:m,sessionId:null};if(s.startsWith("/profiles/")&&s.length>10){const p=s.slice(10).split("/"),j=decodeURIComponent(p[0]);if(j.includes(":"))return{view:"network",profile:"all",tab:"borrowed",mcpScope:m,sessionId:null};let k=g;if(p[1]){const b=decodeURIComponent(p[1]);k=b==="telemetry"?"quota":b}return{view:"profile-detail",profile:j,tab:k,mcpScope:m,sessionId:null}}if(c)return{view:"session-detail",profile:f||"all",tab:g,mcpScope:m,sessionId:c};if(u==="sessions")return{view:"sessions",profile:f||"all",tab:g,mcpScope:m,sessionId:null};if(u==="mcp")return{view:"mcp",profile:f||"all",tab:g,mcpScope:m!=="all"?m:f&&f!=="all"?f:"all",sessionId:null};if(f&&f!=="all")return{view:"profile-detail",profile:f,tab:g,mcpScope:m,sessionId:null}}catch{}return{view:"overview",profile:"all",tab:"quota",mcpScope:"all",sessionId:null}}function bw(){const s=h.useMemo(()=>mp(),[]),[r,c]=h.useState(s.view),[u,f]=h.useState(s.view==="network"),[d,m]=h.useState(s.profile),[g,p]=h.useState(s.mcpScope),[j,k]=h.useState(s.tab),[b,v]=h.useState(null),[z,E]=h.useState(s.sessionId),[$,B]=h.useState(s.agentHint||null),[F,Me]=h.useState("overview"),[te,C]=h.useState([]),[ie,de]=h.useState({version:"v0.12.0",status:"healthy",active_sessions:2,total_profiles:4,terminals_available:["tmux","wezterm","iterm2","terminal"]}),[W,Ee]=h.useState(!0),[ne,Ce]=h.useState(null),[pe,V]=h.useState(!1),[Ie,he]=h.useState(!1),ue=h.useCallback((xe,We,gn,Qt,mn,Re=!1)=>{try{let Ue="/";const un=new URLSearchParams;xe==="session-detail"&&mn?(Ue=`/sessions/${encodeURIComponent(mn)}`,We&&We!=="all"&&un.set("profile",We)):xe==="profile-detail"&&We&&We!=="all"?gn&&gn!=="quota"?Ue=`/profiles/${encodeURIComponent(We)}/${encodeURIComponent(gn)}`:Ue=`/profiles/${encodeURIComponent(We)}`:xe==="sessions"?(Ue="/sessions",We&&We!=="all"&&un.set("profile",We)):xe==="mcp"?(Ue="/mcp",Qt&&Qt!=="all"&&un.set("scope",Qt)):xe==="network"&&(Ue="/network");const ul=un.toString(),Cn=ul?`${Ue}?${ul}`:Ue,bn=`${window.location.pathname}${window.location.search}`;Cn!==bn&&(Re?window.history.replaceState({view:xe,profile:We,tab:gn,mcpScope:Qt,sessionId:mn},"",Cn):window.history.pushState({view:xe,profile:We,tab:gn,mcpScope:Qt,sessionId:mn},"",Cn))}catch{}},[]);h.useEffect(()=>{const xe=()=>{const We=mp();We.view==="network"&&f(!0),c(We.view),m(We.profile),k(We.tab),p(We.mcpScope),We.sessionId?(E(We.sessionId),B(We.agentHint||null),(b==null?void 0:b.id)!==We.sessionId&&v(null)):(v(null),E(null),B(null))};return window.addEventListener("popstate",xe),()=>window.removeEventListener("popstate",xe)},[b]),h.useEffect(()=>{z&&(!b||b.id!==z)&&__().then(xe=>{const We=xe.find(gn=>gn.id===z||($?gn.agent===$&&gn.id.startsWith(z):!1)||gn.id.startsWith(z));v(We||{id:z,agent:$||"claude",profile:d!=="all"?d:"work",title:`Session ${z.slice(0,8)}`,cwd:"",goal:"",turns:0,updated_at:new Date().toISOString(),is_active:!1})}).catch(()=>{v({id:z,agent:$||"claude",profile:d!=="all"?d:"work",title:`Session ${z.slice(0,8)}`,cwd:"",goal:"",turns:0,updated_at:new Date().toISOString(),is_active:!1})})},[z,$,b,d]);const G=h.useCallback(async()=>{try{const xe=await Db();C(xe)}catch{}},[]),we=h.useCallback(async()=>{try{const xe=await Bb();Ce(xe)}catch{}},[]);h.useEffect(()=>{async function xe(){try{const gn=await Ob();de(gn),Ee(!0)}catch{Ee(!1)}}xe(),we(),G(),__().catch(()=>{});const We=setInterval(()=>{xe(),we(),G(),__().catch(()=>{})},3e4);return()=>clearInterval(We)},[we,G]);const fe=xe=>{v(null),E(null),xe==="overview"?(m("all"),c("overview"),ue("overview","all","quota","all",null)):xe==="sessions"?(m("all"),c("sessions"),ue("sessions","all","quota","all",null)):xe==="mcp"?(p("all"),c("mcp"),ue("mcp","all","quota","all",null)):xe==="network"&&(f(!0),c("network"),ue("network","all","quota","all",null))},Oe=xe=>{if(v(null),E(null),xe.includes(":")){fe("network");return}xe==="all"?(m("all"),c("overview"),ue("overview","all","quota","all",null)):(m(xe),c("profile-detail"),ue("profile-detail",xe,j,"all",null))},Se=xe=>{v(null),E(null),m(xe),c("sessions"),ue("sessions",xe,"quota","all",null)},Qe=xe=>{v(null),E(null),p(xe),c("mcp"),ue("mcp","all","quota",xe,null)},Ut=xe=>{Me(r),v(xe),E(xe.id),c("session-detail"),ue("session-detail",xe.profile||d,j,g,xe.id)},Ye=()=>{v(null),E(null),F==="profile-detail"&&d&&d!=="all"?(c("profile-detail"),ue("profile-detail",d,j,"all",null)):(c("sessions"),ue("sessions",d,"quota","all",null))},Ne=xe=>{k(xe),r==="profile-detail"&&ue("profile-detail",d,xe,"all",null,!0)},Xt=()=>{switch(r){case"overview":return o.jsx("div",{className:"flex items-center gap-1.5 text-sm font-mono",children:o.jsx("span",{className:"text-foreground font-semibold",children:"All Agent Profiles"})});case"profile-detail":return o.jsxs("div",{className:"flex items-center gap-1.5 text-sm font-mono",children:[o.jsx("button",{type:"button",onClick:()=>fe("overview"),className:"text-muted-foreground hover:text-foreground cursor-pointer transition-colors",children:"Profiles"}),o.jsx(Ji,{className:"h-3.5 w-3.5 text-muted-foreground/60"}),o.jsx("span",{className:"text-foreground font-semibold",children:d})]});case"sessions":return o.jsxs("div",{className:"flex items-center gap-1.5 text-sm font-mono",children:[o.jsx("span",{className:"text-foreground font-semibold",children:"Sessions Explorer"}),d!=="all"&&o.jsxs(o.Fragment,{children:[o.jsx(Ji,{className:"h-3.5 w-3.5 text-muted-foreground/60"}),o.jsx("span",{className:"text-muted-foreground",children:d})]})]});case"session-detail":{const xe=(b==null?void 0:b.agent)||$;return o.jsxs("div",{className:"flex items-center gap-1.5 text-sm font-mono",children:[o.jsx("button",{type:"button",onClick:Ye,className:"text-muted-foreground hover:text-foreground cursor-pointer transition-colors",children:"Sessions"}),xe&&o.jsxs(o.Fragment,{children:[o.jsx(Ji,{className:"h-3.5 w-3.5 text-muted-foreground/60"}),o.jsx("span",{className:"text-muted-foreground",children:xe})]}),o.jsx(Ji,{className:"h-3.5 w-3.5 text-muted-foreground/60"}),o.jsx("span",{className:"text-foreground font-semibold",children:b?b.id.slice(0,8):z?z.slice(0,8):"Detail"})]})}case"mcp":return o.jsxs("div",{className:"flex items-center gap-1.5 text-sm font-mono",children:[o.jsx("span",{className:"text-foreground font-semibold",children:"MCP Catalog"}),g!=="all"&&o.jsxs(o.Fragment,{children:[o.jsx(Ji,{className:"h-3.5 w-3.5 text-muted-foreground/60"}),o.jsx("span",{className:"text-muted-foreground",children:g})]})]});case"network":return o.jsx("div",{className:"flex items-center gap-1.5 text-sm font-mono",children:o.jsx("span",{className:"text-foreground font-semibold",children:"P2P Network Sharing"})})}};return o.jsxs("div",{className:"flex h-[100dvh] w-full bg-background text-foreground overflow-hidden font-sans",children:[o.jsx(b1,{currentView:r,selectedProfile:d,profiles:te,daemonInfo:ne,version:ie.version,isOnline:W,onSelectView:fe,onSelectProfile:Oe,onOpenAddProfile:()=>he(!0),onOpenDaemonModal:()=>V(!0)}),o.jsxs("div",{className:"flex-1 flex flex-col min-w-0 h-full overflow-hidden bg-background",children:[o.jsx("header",{className:"h-14 border-b border-border bg-background/90 backdrop-blur-md px-4 sm:px-6 flex items-center justify-between shrink-0",children:o.jsx("div",{className:"flex items-center gap-2.5",children:Xt()})}),o.jsx("main",{className:"flex-1 overflow-y-auto px-4 sm:px-6 py-4 sm:py-6 min-h-0 bg-background",children:o.jsx("div",{className:"max-w-7xl mx-auto",children:o.jsx(q1,{children:o.jsxs(h.Suspense,{fallback:r==="network"?o.jsx($1,{}):r==="profile-detail"?o.jsx(Y1,{profileName:d}):r==="mcp"?o.jsx(H1,{}):r==="session-detail"?o.jsx(A0,{targetSessionId:z,onBack:Ye}):r==="overview"?o.jsx(U1,{}):o.jsx(L1,{}),children:[r==="overview"&&o.jsx(_p,{selectedProfile:"all",profiles:te,onSelectProfile:Oe,onNavigateToSessions:Se,onNavigateToMcp:Qe,onProfilesLoaded:C}),r==="profile-detail"&&o.jsx(_p,{selectedProfile:d,profiles:te,activeTab:j,onTabChange:Ne,onSelectProfile:Oe,onNavigateToSessions:Se,onNavigateToMcp:Qe,onProfilesLoaded:C,onViewSession:Ut}),o.jsx("div",{className:r==="sessions"?"block":"hidden",children:o.jsx(gw,{selectedProfile:d,onSelectProfile:m,profiles:te,onViewSession:Ut})}),r==="session-detail"&&(b?o.jsx(pw,{session:b,onBack:Ye},b.id):o.jsx(A0,{targetSessionId:z,onBack:Ye})),r==="mcp"&&o.jsx(yw,{selectedScope:g,onSelectScope:xe=>{p(xe),ue("mcp","all","quota",xe,null,!0)},profiles:te}),o.jsx("div",{className:r==="network"?"block":"hidden",children:(u||r==="network")&&o.jsx(xw,{profiles:te})})]})})})})]}),o.jsx(T1,{open:pe,onOpenChange:V,daemonInfo:ne,onRefreshDaemon:we}),o.jsx(R1,{open:Ie,onOpenChange:he,onProfileCreated:()=>{G()}}),o.jsx(D1,{}),o.jsx(mw,{endpoint:"http://localhost:4747",appName:"AIM"})]})}Rb.createRoot(document.getElementById("root")).render(o.jsx(kb.StrictMode,{children:o.jsx(bw,{})}));export{R1 as A,$o as B,Hw as C,Dp as D,qw as E,Iw as F,Q_ as G,$w as H,Pu as I,Uw as J,Dw as K,H1 as M,Y1 as P,Bp as S,B1 as T,$_ as a,U_ as b,Y_ as c,q_ as d,H_ as e,Du as f,Pt as g,Db as h,__ as i,Cw as j,Tw as k,U1 as l,Nw as m,jw as n,Ew as o,Mw as p,w1 as q,kw as r,Ow as s,Bw as t,X_ as u,Lw as v,Yw as w,zw as x,Rw as y,Aw as z};

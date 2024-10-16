"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[7921],{1488:(e,n,i)=>{i.r(n),i.d(n,{assets:()=>c,contentTitle:()=>a,default:()=>h,frontMatter:()=>t,metadata:()=>r,toc:()=>d});var o=i(5893),s=i(1151);const t={sidebar_position:4,title:"Equivocation governance proposal"},a="ADR 003: Equivocation governance proposal",r={id:"adrs/adr-003-equivocation-gov-proposal",title:"Equivocation governance proposal",description:"Changelog",source:"@site/versioned_docs/version-v5.2.0/adrs/adr-003-equivocation-gov-proposal.md",sourceDirName:"adrs",slug:"/adrs/adr-003-equivocation-gov-proposal",permalink:"/interchain-security/v5.2.0/adrs/adr-003-equivocation-gov-proposal",draft:!1,unlisted:!1,tags:[],version:"v5.2.0",sidebarPosition:4,frontMatter:{sidebar_position:4,title:"Equivocation governance proposal"},sidebar:"tutorialSidebar",previous:{title:"Jail Throttling",permalink:"/interchain-security/v5.2.0/adrs/adr-002-throttle"},next:{title:"Cryptographic verification of equivocation evidence",permalink:"/interchain-security/v5.2.0/adrs/adr-005-cryptographic-equivocation-verification"}},c={},d=[{value:"Changelog",id:"changelog",level:2},{value:"Status",id:"status",level:2},{value:"Context",id:"context",level:2},{value:"Decision",id:"decision",level:2},{value:"Consequences",id:"consequences",level:2},{value:"Positive",id:"positive",level:3},{value:"Negative",id:"negative",level:3},{value:"Neutral",id:"neutral",level:3},{value:"References",id:"references",level:2}];function l(e){const n={a:"a",code:"code",em:"em",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,s.a)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(n.h1,{id:"adr-003-equivocation-governance-proposal",children:"ADR 003: Equivocation governance proposal"}),"\n",(0,o.jsx)(n.h2,{id:"changelog",children:"Changelog"}),"\n",(0,o.jsxs)(n.ul,{children:["\n",(0,o.jsx)(n.li,{children:"2023-02-06: Initial draft"}),"\n",(0,o.jsx)(n.li,{children:"2023-11-30: Change status to deprecated"}),"\n"]}),"\n",(0,o.jsx)(n.h2,{id:"status",children:"Status"}),"\n",(0,o.jsx)(n.p,{children:"Deprecated"}),"\n",(0,o.jsx)(n.h2,{id:"context",children:"Context"}),"\n",(0,o.jsxs)(n.p,{children:[(0,o.jsx)(n.strong,{children:"Note:"})," ADR deprecated as the equivocation proposal was removed by the\ncryptographic verification of equivocation feature\n(see ",(0,o.jsx)(n.a,{href:"/interchain-security/v5.2.0/adrs/adr-005-cryptographic-equivocation-verification",children:"ADR-005"})," and\n",(0,o.jsx)(n.a,{href:"/interchain-security/v5.2.0/adrs/adr-013-equivocation-slashing",children:"ADR-013"}),")."]}),"\n",(0,o.jsx)(n.p,{children:"We want to limit the possibilities of a consumer chain to execute actions on the provider chain to maintain and ensure optimum security of the provider chain."}),"\n",(0,o.jsx)(n.p,{children:"For instance, a malicious consumer consumer chain can send slash packet to the provider chain, which will slash a validator without the need of providing an evidence."}),"\n",(0,o.jsx)(n.h2,{id:"decision",children:"Decision"}),"\n",(0,o.jsx)(n.p,{children:"To protect against a malicious consumer chain, slash packets unrelated to downtime are ignored by the provider chain. Thus, an other mechanism is required to punish validators that have committed a double-sign on a consumer chain."}),"\n",(0,o.jsxs)(n.p,{children:["A new kind of governance proposal is added to the ",(0,o.jsx)(n.code,{children:"provider"})," module, allowing to slash and tombstone a validator for double-signing in case of any harmful action on the consumer chain."]}),"\n",(0,o.jsxs)(n.p,{children:["If such proposal passes, the proposal handler delegates to the ",(0,o.jsx)(n.code,{children:"evidence"})," module to process the equivocation. This module ensures the evidence isn\u2019t too old, or else ignores it (see ",(0,o.jsx)(n.a,{href:"https://github.com/cosmos/cosmos-sdk/blob/21021b837882d1d40f1d79bcbc4fad2e79a3fefe/x/evidence/keeper/infraction.go#L54-L62",children:"code"}),"). ",(0,o.jsx)(n.em,{children:"Too old"})," is determined by 2 consensus params :"]}),"\n",(0,o.jsxs)(n.ul,{children:["\n",(0,o.jsxs)(n.li,{children:[(0,o.jsx)(n.code,{children:"evidence.max_age_duration"})," number of nanoseconds before an evidence is considered too old"]}),"\n",(0,o.jsxs)(n.li,{children:[(0,o.jsx)(n.code,{children:"evidence.max_age_numblocks"})," number of blocks before an evidence is considered too old."]}),"\n"]}),"\n",(0,o.jsx)(n.p,{children:"On the hub, those parameters are equals to"}),"\n",(0,o.jsx)(n.pre,{children:(0,o.jsx)(n.code,{className:"language-json",children:'// From https://cosmos-rpc.polkachu.com/consensus_params?height=13909682\n(...)\n"evidence": {\n\t"max_age_num_blocks": "1000000",\n\t"max_age_duration": "172800000000000",\n  (...)\n},\n(...)\n'})}),"\n",(0,o.jsxs)(n.p,{children:["A governance proposal takes 14 days, so those parameters must be big enough so the evidence provided in the proposal is not ignored by the ",(0,o.jsx)(n.code,{children:"evidence"})," module when the proposal passes and is handled by the hub."]}),"\n",(0,o.jsxs)(n.p,{children:["For ",(0,o.jsx)(n.code,{children:"max_age_num_blocks=1M"}),", the parameter is big enough if we consider the hub produces 12k blocks per day (",(0,o.jsx)(n.code,{children:"blocks_per_year/365 = 436,0000/365"}),"). The evidence can be up to 83 days old (",(0,o.jsx)(n.code,{children:"1,000,000/12,000"}),") and not be ignored."]}),"\n",(0,o.jsxs)(n.p,{children:["For ",(0,o.jsx)(n.code,{children:"max_age_duration=172,800,000,000,000"}),", the parameter is too low, because the value is in nanoseconds so it\u2019s 2 days. Fortunately the condition that checks those 2 parameters uses a ",(0,o.jsx)(n.strong,{children:"AND"}),", so if ",(0,o.jsx)(n.code,{children:"max_age_num_blocks"})," condition passes, the evidence won\u2019t be ignored."]}),"\n",(0,o.jsx)(n.h2,{id:"consequences",children:"Consequences"}),"\n",(0,o.jsx)(n.h3,{id:"positive",children:"Positive"}),"\n",(0,o.jsxs)(n.ul,{children:["\n",(0,o.jsx)(n.li,{children:"Remove the possibility from a malicious consumer chain to \u201cattack\u201d the provider chain by slashing/jailing validators."}),"\n",(0,o.jsx)(n.li,{children:"Provide a more acceptable implementation for the validator community."}),"\n"]}),"\n",(0,o.jsx)(n.h3,{id:"negative",children:"Negative"}),"\n",(0,o.jsxs)(n.ul,{children:["\n",(0,o.jsx)(n.li,{children:"Punishment action of double-signing isn\u2019t \u201cautomated\u201d, a governance proposal is required which takes more time."}),"\n",(0,o.jsx)(n.li,{children:"You need to pay 250ATOM to submit an equivocation evidence."}),"\n"]}),"\n",(0,o.jsx)(n.h3,{id:"neutral",children:"Neutral"}),"\n",(0,o.jsx)(n.h2,{id:"references",children:"References"}),"\n",(0,o.jsxs)(n.ul,{children:["\n",(0,o.jsxs)(n.li,{children:["PR that ignores non downtime slash packet : ",(0,o.jsx)(n.a,{href:"https://github.com/cosmos/interchain-security/pull/692",children:"https://github.com/cosmos/interchain-security/pull/692"})]}),"\n",(0,o.jsxs)(n.li,{children:["PR that adds the governance slash proposal: ",(0,o.jsx)(n.a,{href:"https://github.com/cosmos/interchain-security/pull/703",children:"https://github.com/cosmos/interchain-security/pull/703"})]}),"\n"]})]})}function h(e={}){const{wrapper:n}={...(0,s.a)(),...e.components};return n?(0,o.jsx)(n,{...e,children:(0,o.jsx)(l,{...e})}):l(e)}},1151:(e,n,i)=>{i.d(n,{Z:()=>r,a:()=>a});var o=i(7294);const s={},t=o.createContext(s);function a(e){const n=o.useContext(t);return o.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function r(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:a(e.components),o.createElement(t.Provider,{value:n},e.children)}}}]);
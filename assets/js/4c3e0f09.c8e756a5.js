"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[5637],{8064:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>d,contentTitle:()=>i,default:()=>h,frontMatter:()=>a,metadata:()=>o,toc:()=>c});var r=s(5893),t=s(1151);const a={sidebar_position:3,title:"Key Assignment"},i="ADR 001: Key Assignment",o={id:"adrs/adr-001-key-assignment",title:"Key Assignment",description:"Changelog",source:"@site/versioned_docs/version-v6.1.0/adrs/adr-001-key-assignment.md",sourceDirName:"adrs",slug:"/adrs/adr-001-key-assignment",permalink:"/interchain-security/v6.1.0/adrs/adr-001-key-assignment",draft:!1,unlisted:!1,tags:[],version:"v6.1.0",sidebarPosition:3,frontMatter:{sidebar_position:3,title:"Key Assignment"},sidebar:"tutorialSidebar",previous:{title:"Pause validator unbonding during equivocation proposal",permalink:"/interchain-security/v6.1.0/adrs/adr-007-pause-unbonding-on-eqv-prop"},next:{title:"Jail Throttling",permalink:"/interchain-security/v6.1.0/adrs/adr-002-throttle"}},d={},c=[{value:"Changelog",id:"changelog",level:2},{value:"Status",id:"status",level:2},{value:"Context",id:"context",level:2},{value:"Decision",id:"decision",level:2},{value:"State required",id:"state-required",level:3},{value:"Protocol overview",id:"protocol-overview",level:3},{value:"Consequences",id:"consequences",level:2},{value:"Positive",id:"positive",level:3},{value:"Negative",id:"negative",level:3},{value:"Neutral",id:"neutral",level:3},{value:"References",id:"references",level:2}];function l(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,t.a)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(n.h1,{id:"adr-001-key-assignment",children:"ADR 001: Key Assignment"}),"\n",(0,r.jsx)(n.h2,{id:"changelog",children:"Changelog"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsx)(n.li,{children:"2022-12-01: Initial Draft"}),"\n",(0,r.jsx)(n.li,{children:"2024-03-01: Updated to take into account they key-assigment-replacement deprecation."}),"\n"]}),"\n",(0,r.jsx)(n.h2,{id:"status",children:"Status"}),"\n",(0,r.jsx)(n.p,{children:"Accepted"}),"\n",(0,r.jsx)(n.h2,{id:"context",children:"Context"}),"\n",(0,r.jsx)(n.p,{children:"KeyAssignment is the name of the feature that allows validator operators to use different consensus keys for each consumer chain validator node that they operate."}),"\n",(0,r.jsx)(n.h2,{id:"decision",children:"Decision"}),"\n",(0,r.jsxs)(n.p,{children:["It is possible to change the keys at any time by submitting a transaction (i.e., ",(0,r.jsx)(n.code,{children:"MsgAssignConsumerKey"}),")."]}),"\n",(0,r.jsx)(n.h3,{id:"state-required",children:"State required"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.code,{children:"ValidatorConsumerPubKey"})," - Stores the validator assigned keys for every consumer chain."]}),"\n"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"ConsumerValidatorsBytePrefix | len(chainID) | chainID | providerConsAddress -> consumerKey\n"})}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.code,{children:"ValidatorByConsumerAddr"})," - Stores the mapping from validator addresses on consumer chains to validator addresses on the provider chain. Needed for the consumer initiated slashing sub-protocol."]}),"\n"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"ValidatorsByConsumerAddrBytePrefix | len(chainID) | chainID | consumerConsAddress -> providerConsAddress\n"})}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.code,{children:"ConsumerAddrsToPrune"})," - Stores the mapping from VSC ids to consumer validators addresses. Needed for pruning ",(0,r.jsx)(n.code,{children:"ValidatorByConsumerAddr"}),"."]}),"\n"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"ConsumerAddrsToPruneBytePrefix | len(chainID) | chainID | vscID -> []consumerConsAddresses\n"})}),"\n",(0,r.jsx)(n.h3,{id:"protocol-overview",children:"Protocol overview"}),"\n",(0,r.jsxs)(n.p,{children:["On receiving a ",(0,r.jsx)(n.code,{children:"MsgAssignConsumerKey(chainID, providerAddr, consumerKey)"})," message:"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"// get validator from staking module  \nvalidator, found := stakingKeeper.GetValidator(providerAddr)\nif !found {\n    return ErrNoValidatorFound\n}\nproviderConsAddr := validator.GetConsAddr()\n\n// make sure consumer key is not in use\nconsumerAddr := utils.TMCryptoPublicKeyToConsAddr(consumerKey)\nif _, found := GetValidatorByConsumerAddr(ChainID, consumerAddr); found {\n    return ErrInvalidConsumerConsensusPubKey\n}\n\n// check whether the consumer chain is already registered\n// i.e., a client to the consumer was already created\nif _, consumerRegistered := GetConsumerClientId(chainID); consumerRegistered {\n    // get the previous key assigned for this validator on this consumer chain\n    oldConsumerKey, found := GetValidatorConsumerPubKey(chainID, providerConsAddr)\n    if found {\n        // mark this old consumer key as prunable once the VSCMaturedPacket\n        // for the current VSC ID is received\n        oldConsumerAddr := utils.TMCryptoPublicKeyToConsAddr(oldConsumerKey)\n        vscID := GetValidatorSetUpdateId()\n        AppendConsumerAddrsToPrune(chainID, vscID, oldConsumerAddr)\n    }\n} else {\n    // if the consumer chain is not registered, then remove the previous reverse mapping\n    if oldConsumerKey, found := GetValidatorConsumerPubKey(chainID, providerConsAddr); found {\n        oldConsumerAddr := utils.TMCryptoPublicKeyToConsAddr(oldConsumerKey)\n        DeleteValidatorByConsumerAddr(chainID, oldConsumerAddr)\n    }\n}\n\n\n// set the mapping from this validator's provider address to the new consumer key\nSetValidatorConsumerPubKey(chainID, providerConsAddr, consumerKey)\n\n// set the reverse mapping: from this validator's new consensus address \n// on the consumer to its consensus address on the provider\nSetValidatorByConsumerAddr(chainID, consumerAddr, providerConsAddr)\n"})}),"\n",(0,r.jsxs)(n.p,{children:["When a new consumer chain is registered, i.e., a client to the consumer chain is created, the provider constructs the consumer CCV module part of the genesis state (see ",(0,r.jsx)(n.code,{children:"MakeConsumerGenesis"}),")."]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"func (k Keeper) MakeConsumerGenesis(chainID string) (gen consumertypes.GenesisState, nextValidatorsHash []byte, err error) {\n    // ...\n    // get initial valset from the staking module\n    var updates []abci.ValidatorUpdate{}\n    stakingKeeper.IterateLastValidatorPowers(func(providerAddr sdk.ValAddress, power int64) (stop bool) {\n        validator := stakingKeeper.GetValidator(providerAddr)\n        providerKey := validator.TmConsPublicKey()\n\t\tupdates = append(updates, abci.ValidatorUpdate{PubKey: providerKey, Power: power})\n\t\treturn false\n\t})\n\n    // applies the key assignment to the initial validator\n\tfor i, update := range updates {\n\t\tproviderAddr := utils.TMCryptoPublicKeyToConsAddr(update.PubKey)\n\t\tif consumerKey, found := GetValidatorConsumerPubKey(chainID, providerAddr); found {\n\t\t\tupdates[i].PubKey = consumerKey\n\t\t}\n\t}\n    gen.InitialValSet = updates\n\n    // get a hash of the consumer validator set from the update\n\tupdatesAsValSet := tendermint.PB2TM.ValidatorUpdates(updates)\n\thash := tendermint.NewValidatorSet(updatesAsValSet).Hash()\n\n\treturn gen, hash, nil\n}\n"})}),"\n",(0,r.jsxs)(n.p,{children:["Note that key assignment works hand-in-hand with ",(0,r.jsx)(n.a,{href:"https://github.com/cosmos/interchain-security/blob/main/docs/docs/adrs/adr-014-epochs.md",children:"epochs"}),".\nFor each consumer chain, we store the consumer validator set that is currently (i.e., in this epoch) validating the consumer chain.\nSpecifically, for each validator in the set we store among others, the public key that it is using on the consumer chain during the current (i.e., ongoing) epoch.\nAt the end of every epoch, if there were validator set changes on the provider, then for every consumer chain, we construct a ",(0,r.jsx)(n.code,{children:"VSCPacket"}),"\nwith all the validator updates and add it to the list of ",(0,r.jsx)(n.code,{children:"PendingVSCPacket"}),"s. We compute the validator updates needed by a consumer chain by\ncomparing the stored list of consumer validators with the current bonded validators on the provider, with something similar to this:"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"// get the valset that has been validating the consumer chain during this epoch \ncurrentValidators := GetConsumerValSet(consumerChain)\n// generate the validator updates needed to be sent through a `VSCPacket` by comparing the current validators \n// in the epoch with the latest bonded validators\nvalUpdates := DiffValidators(currentValidators, stakingmodule.GetBondedValidators())\n// update the current validators set for the upcoming epoch to be the latest bonded validators instead\nSetConsumerValSet(stakingmodule.GetBondedValidators())\n"})}),"\n",(0,r.jsxs)(n.p,{children:["where ",(0,r.jsx)(n.code,{children:"DiffValidators"})," internally checks if the consumer public key for a validator has changed since the last\nepoch and if so generates a validator update. This way, a validator can change its consumer public key for a consumer\nchain an arbitrary amount of times and only the last set consumer public key would be taken into account."]}),"\n",(0,r.jsxs)(n.p,{children:["On receiving a ",(0,r.jsx)(n.code,{children:"SlashPacket"})," from a consumer chain with id ",(0,r.jsx)(n.code,{children:"chainID"})," for a infraction of a validator ",(0,r.jsx)(n.code,{children:"data.Validator"}),":"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"func HandleSlashPacket(chainID string, data ccv.SlashPacketData) (success bool, err error) {\n    // ...\n    // the slash packet validator address may be known only on the consumer chain;\n\t// in this case, it must be mapped back to the consensus address on the provider chain\n    consumerAddr := sdk.ConsAddress(data.Validator.Address)\n    providerAddr, found := GetValidatorByConsumerAddr(chainID, consumerAddr)\n    if !found {\n        // the validator has the same key on the consumer as on the provider\n        providerAddr = consumerAddr\n    }\n    // ...\n}\n"})}),"\n",(0,r.jsxs)(n.p,{children:["On receiving a ",(0,r.jsx)(n.code,{children:"VSCMatured"}),":"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"func OnRecvVSCMaturedPacket(packet channeltypes.Packet, data ccv.VSCMaturedPacketData) exported.Acknowledgement {\n    // ...\n    // prune previous consumer validator address that are no longer needed\n    consumerAddrs := GetConsumerAddrsToPrune(chainID, data.ValsetUpdateId)\n\tfor _, addr := range consumerAddrs {\n\t\tDeleteValidatorByConsumerAddr(chainID, addr)\n\t}\n\tDeleteConsumerAddrsToPrune(chainID, data.ValsetUpdateId)\n    // ...\n}\n"})}),"\n",(0,r.jsx)(n.p,{children:"On stopping a consumer chain:"}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-golang",children:"func (k Keeper) StopConsumerChain(ctx sdk.Context, chainID string, closeChan bool) (err error) {\n    // ...\n    // deletes all the state needed for key assignments on this consumer chain\n    // ...\n}\n"})}),"\n",(0,r.jsx)(n.h2,{id:"consequences",children:"Consequences"}),"\n",(0,r.jsx)(n.h3,{id:"positive",children:"Positive"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsx)(n.li,{children:"Validators can use different consensus keys on the consumer chains."}),"\n"]}),"\n",(0,r.jsx)(n.h3,{id:"negative",children:"Negative"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsx)(n.li,{children:"None"}),"\n"]}),"\n",(0,r.jsx)(n.h3,{id:"neutral",children:"Neutral"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["The consensus state necessary to create a client to the consumer chain must use the hash returned by the ",(0,r.jsx)(n.code,{children:"MakeConsumerGenesis"})," method as the ",(0,r.jsx)(n.code,{children:"nextValsHash"}),"."]}),"\n",(0,r.jsxs)(n.li,{children:["The consumer chain can no longer check the initial validator set against the consensus state on ",(0,r.jsx)(n.code,{children:"InitGenesis"}),"."]}),"\n"]}),"\n",(0,r.jsx)(n.h2,{id:"references",children:"References"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsx)(n.li,{children:(0,r.jsx)(n.a,{href:"https://github.com/cosmos/interchain-security/issues/26",children:"Key assignment issue"})}),"\n"]})]})}function h(e={}){const{wrapper:n}={...(0,t.a)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(l,{...e})}):l(e)}},1151:(e,n,s)=>{s.d(n,{Z:()=>o,a:()=>i});var r=s(7294);const t={},a=r.createContext(t);function i(e){const n=r.useContext(a);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function o(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:i(e.components),r.createElement(a.Provider,{value:n},e.children)}}}]);
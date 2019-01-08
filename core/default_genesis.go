package core

import (
	"encoding/json"
	"github.com/matrix/go-matrix/base58"
	"github.com/matrix/go-matrix/common"
)

var (
	DefaultJson = `{
    "nettopology":{
        "Type":0,
        "NetTopologyData":[
            {
                "Account":"MAN.44EuST4f2vLeEMw2bsMWmBYqLMBhi",
                "Position":8192
            },
            {
                "Account":"MAN.3t9Ser2UjrXRT6erVKytjHtT4ohdX",
                "Position":8193
            },
            {
                "Account":"MAN.2MSWsig8iv45CDTrPn1XMzvZnR52v",
                "Position":8194
            },
            {
                "Account":"MAN.EXnVsEqjyPLHZySL9y53WgXjiFmN",
                "Position":8195
            },
            {
                "Account":"MAN.2QR75feL9KBfaezJwhuM2VCPpqkyT",
                "Position":8196
            },
            {
                "Account":"MAN.95Tro9wLULb6rNWNCT6QDwVBiDds",
                "Position":8197
            },
            {
                "Account":"MAN.375qsgtLc25bJv2Cf9ffnEZPDhyvd",
                "Position":8198
            },
            {
                "Account":"MAN.2RwBWjBykMiGRu7STzYdNzXUMyh8z",
                "Position":8199
            },
            {
                "Account":"MAN.2zjKitA4uydg5kSZLjtSNHzDtx6k8",
                "Position":8200
            },
            {
                "Account":"MAN.4kaowsz37i1WRrPrkg4g8qYQHtJ7",
                "Position":8201
            },
            {
                "Account":"MAN.3B1wx2wo5anTTMjnyWXAA4FQL5opx",
                "Position":8202
            },
            {
                "Account":"MAN.3pm3iXgrY9SYhhc9aS6DQCp4qj66t",
                "Position":8203
            },
            {
                "Account":"MAN.4M8Svax1v6yGitwB4E2ueBtw2TesK",
                "Position":8204
            },
            {
                "Account":"MAN.3abRtpG81TPY8YwenowwgmJ8SuLpy",
                "Position":8205
            },
            {
                "Account":"MAN.2pqHNWtbpKKaLU71RY6TZeaQjjQnR",
                "Position":8206
            },
            {
                "Account":"MAN.3Wf4yzVkbou1bFzkpbJUdu5rqq1se",
                "Position":8207
            },
            {
                "Account":"MAN.U5mNRj4q7jQzbVE4tWcjsvSXTseS",
                "Position":8208
            },
            {
                "Account":"MAN.2ZPbpaCyuEEf1WR3C6dCCFFqnH25J",
                "Position":8209
            },
            {
                "Account":"MAN.3ByNBjw4E7gcxD3uGKtAZYLYUD5xi",
                "Position":8210
            },
            {
                "Account":"MAN.2tSj5kjwwaiTPt4XnHAaEWBYBo9gK",
                "Position":0
            },
            {
                "Account":"MAN.4ZnJrUuM2bfFmdUaivrqGio28hZwd",
                "Position":1
            },
            {
                "Account":"MAN.49fxdGyiWPQ3evpMCNChqvCq3qzMC",
                "Position":2
            },
            {
                "Account":"MAN.3Ugik7ZsLoaNgX51kCJEWz1ZjxQgW",
                "Position":3
            },
            {
                "Account":"MAN.kwPCJkajT2op7rVgYKDqcQu2KEQn",
                "Position":4
            },
            {
                "Account":"MAN.ksFr4mKPfZhm2PrFdEUSoLoDsKAZ",
                "Position":12288
            },
            {
                "Account":"MAN.42bUyszBXL3feeDHztWMiUJCRzBRP",
                "Position":12289
            },
            {
                "Account":"MAN.2NVAVDc7AJGNP3Ghwfv8dz59kUvjM",
                "Position":12290
            },
            {
                "Account":"MAN.e33HPpmmZC98ADkZUXigS1nDFfaA",
                "Position":12291
            }
        ]
    },
    "alloc":{
        "MAN.1111111111111111111B8":{
            "storage":{
                "0x0000000000000000000000000000000000000000000000000000000a444e554d":"0x000000000000000000000000000000000000000000000000000000000000001c",
                "0x0000000000000000000000000000000000000000000a44490000000000000000":"0x000000000000000000000000db588e42D894EDE75000860fC0F4D969393ac514",
                "0x0000000000000000000000000000000000000000000a44490000000000000001":"0x000000000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA4",
                "0x0000000000000000000000000000000000000000000a44490000000000000002":"0x000000000000000000000000611318F1430AD50c540677089053dD7123e3B9b1",
                "0x0000000000000000000000000000000000000000000a44490000000000000003":"0x00000000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c",
                "0x0000000000000000000000000000000000000000000a44490000000000000004":"0x00000000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE",
                "0x0000000000000000000000000000000000000000000a44490000000000000005":"0x00000000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca30",
                "0x0000000000000000000000000000000000000000000a44490000000000000006":"0x00000000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf7385",
                "0x0000000000000000000000000000000000000000000a44490000000000000007":"0x00000000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc47",
                "0x0000000000000000000000000000000000000000000a44490000000000000008":"0x0000000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee1",
                "0x0000000000000000000000000000000000000000000a44490000000000000009":"0x00000000000000000000000004A48442762386D1954895a3d1977457145e8Ca7",
                "0x0000000000000000000000000000000000000000000a4449000000000000000a":"0x0000000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E1593",
                "0x0000000000000000000000000000000000000000000a4449000000000000000b":"0x000000000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f",
                "0x0000000000000000000000000000000000000000000a4449000000000000000c":"0x000000000000000000000000f03F2c53784149f95535DBc707277363C6DB45A4",
                "0x0000000000000000000000000000000000000000000a4449000000000000000d":"0x000000000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD7745",
                "0x0000000000000000000000000000000000000000000a4449000000000000000e":"0x00000000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b",
                "0x0000000000000000000000000000000000000000000a4449000000000000000f":"0x000000000000000000000000B442685E3730F1441aE712E4F9E27946b8471902",
                "0x0000000000000000000000000000000000000000000a44490000000000000010":"0x000000000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d01",
                "0x0000000000000000000000000000000000000000000a44490000000000000011":"0x0000000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f746",
                "0x0000000000000000000000000000000000000000000a44490000000000000012":"0x0000000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B997",
                "0x0000000000000000000000000000000000000000000a44490000000000000013":"0x000000000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF",
                "0x0000000000000000000000000000000000000000000a44490000000000000014":"0x000000000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE",
                "0x0000000000000000000000000000000000000000000a44490000000000000015":"0x000000000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E",
                "0x0000000000000000000000000000000000000000000a44490000000000000016":"0x000000000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E469",
                "0x0000000000000000000000000000000000000000000a44490000000000000017":"0x0000000000000000000000003660302D6614EF96578EFC3301cc527054Bc8919",
                "0x0000000000000000000000000000000000000000000a44490000000000000018":"0x0000000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c5",
                "0x0000000000000000000000000000000000000000000a44490000000000000019":"0x000000000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B0",
                "0x0000000000000000000000000000000000000000000a4449000000000000001a":"0x000000000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f0",
                "0x0000000000000000000000000000000000000000000a4449000000000000001b":"0x0000000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd25",
                "0x0000000000000000000000db588e42D894EDE75000860fC0F4D969393ac51444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000611318F1430AD50c540677089053dD7123e3B9b144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca3044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf738544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc4744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000004A48442762386D1954895a3d1977457145e8Ca744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E159344":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000f03F2c53784149f95535DBc707277363C6DB45A444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD774544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000B442685E3730F1441aE712E4F9E27946b847190244":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d0144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f74644":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B99744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF44":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x0000000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE44":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x0000000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E44":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x0000000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E46944":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x00000000000000000000003660302D6614EF96578EFC3301cc527054Bc891944":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x00000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd2544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000db588e42D894EDE75000860fC0F4D969393ac5144e58":"0x000000000000000000000000db588e42D894EDE75000860fC0F4D969393ac514",
                "0x00000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA44e58":"0x000000000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA4",
                "0x00000000000000000000611318F1430AD50c540677089053dD7123e3B9b14e58":"0x000000000000000000000000611318F1430AD50c540677089053dD7123e3B9b1",
                "0x0000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c4e58":"0x00000000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c",
                "0x0000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE4e58":"0x00000000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE",
                "0x0000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca304e58":"0x00000000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca30",
                "0x0000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf73854e58":"0x00000000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf7385",
                "0x0000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc474e58":"0x00000000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc47",
                "0x000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee14e58":"0x0000000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee1",
                "0x0000000000000000000004A48442762386D1954895a3d1977457145e8Ca74e58":"0x00000000000000000000000004A48442762386D1954895a3d1977457145e8Ca7",
                "0x000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E15934e58":"0x0000000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E1593",
                "0x00000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f4e58":"0x000000000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f",
                "0x00000000000000000000f03F2c53784149f95535DBc707277363C6DB45A44e58":"0x000000000000000000000000f03F2c53784149f95535DBc707277363C6DB45A4",
                "0x00000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD77454e58":"0x000000000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD7745",
                "0x0000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b4e58":"0x00000000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b",
                "0x00000000000000000000B442685E3730F1441aE712E4F9E27946b84719024e58":"0x000000000000000000000000B442685E3730F1441aE712E4F9E27946b8471902",
                "0x00000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d014e58":"0x000000000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d01",
                "0x000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f7464e58":"0x0000000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f746",
                "0x000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B9974e58":"0x0000000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B997",
                "0x00000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF4e58":"0x000000000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF",
                "0x00000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE4e58":"0x000000000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE",
                "0x00000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E4e58":"0x000000000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E",
                "0x00000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E4694e58":"0x000000000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E469",
                "0x000000000000000000003660302D6614EF96578EFC3301cc527054Bc89194e58":"0x0000000000000000000000003660302D6614EF96578EFC3301cc527054Bc8919",
                "0x000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c54e58":"0x0000000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c5",
                "0x00000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B04e58":"0x000000000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B0",
                "0x00000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f04e58":"0x000000000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f0",
                "0x000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd254e58":"0x0000000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd25"
            },
            "balance":"2350000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUTz":{
            "balance":"10000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUUs":{
            "balance":"25000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUV2":{
            "balance":"10000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUW7":{
            "balance":"5000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUXN":{
            "balance":"10000000000000000000000000"
        },
        "MAN.4L95KmR3e8eUJvzwK2thft1eKaFYa":{
            "balance":"300000000000000000000000000"
        },
        "MAN.4739r322TyL3xCpbbdohS8NhBgGwi":{
            "balance":"200000000000000000000000000"
        },
        "MAN.2zXWsDtyt7vhVADGTz2yXD6h7WJnF":{
            "balance":"87650000000000000000000000"
        }
    },
    "mstate":{
        "Broadcasts":["MAN.2y5fqzGDWVznvkd49qqWpXiqjcmJF"],
        "curElect":[
            {
                "Account":"MAN.44EuST4f2vLeEMw2bsMWmBYqLMBhi",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3t9Ser2UjrXRT6erVKytjHtT4ohdX",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2MSWsig8iv45CDTrPn1XMzvZnR52v",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.EXnVsEqjyPLHZySL9y53WgXjiFmN",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2QR75feL9KBfaezJwhuM2VCPpqkyT",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.95Tro9wLULb6rNWNCT6QDwVBiDds",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.375qsgtLc25bJv2Cf9ffnEZPDhyvd",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2RwBWjBykMiGRu7STzYdNzXUMyh8z",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2zjKitA4uydg5kSZLjtSNHzDtx6k8",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.4kaowsz37i1WRrPrkg4g8qYQHtJ7",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3B1wx2wo5anTTMjnyWXAA4FQL5opx",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3pm3iXgrY9SYhhc9aS6DQCp4qj66t",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.4M8Svax1v6yGitwB4E2ueBtw2TesK",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3abRtpG81TPY8YwenowwgmJ8SuLpy",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2pqHNWtbpKKaLU71RY6TZeaQjjQnR",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3Wf4yzVkbou1bFzkpbJUdu5rqq1se",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.U5mNRj4q7jQzbVE4tWcjsvSXTseS",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2ZPbpaCyuEEf1WR3C6dCCFFqnH25J",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3ByNBjw4E7gcxD3uGKtAZYLYUD5xi",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2tSj5kjwwaiTPt4XnHAaEWBYBo9gK",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.4ZnJrUuM2bfFmdUaivrqGio28hZwd",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.49fxdGyiWPQ3evpMCNChqvCq3qzMC",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.3Ugik7ZsLoaNgX51kCJEWz1ZjxQgW",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.kwPCJkajT2op7rVgYKDqcQu2KEQn",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.ksFr4mKPfZhm2PrFdEUSoLoDsKAZ",
                "Stock":1,
                "Type":3
            },
            {
                "Account":"MAN.42bUyszBXL3feeDHztWMiUJCRzBRP",
                "Stock":1,
                "Type":3
            },
            {
                "Account":"MAN.2NVAVDc7AJGNP3Ghwfv8dz59kUvjM",
                "Stock":1,
                "Type":3
            },
            {
                "Account":"MAN.e33HPpmmZC98ADkZUXigS1nDFfaA",
                "Stock":1,
                "Type":3
            }
        ],
		"Foundation": "MAN.2zXWsDtyt7vhVADGTz2yXD6h7WJnF",
		"VersionSuperAccounts": [
			"MAN.4739r322TyL3xCpbbdohS8NhBgGwi"
		],
		"BlockSuperAccounts": [
			"MAN.4L95KmR3e8eUJvzwK2thft1eKaFYa"
		],
		"InnerMiners": [
		"MAN.3SPbc3M7bK8zCT8VbvjMGW2eCaBgY"
		],
		"BroadcastInterval": {
			"BCInterval": 100
		},
		"VIPCfg": [
					{
				"MinMoney": 0,
				"InterestRate": 5,
				"ElectUserNum": 0,
				"StockScale": 1000
			},
			{
				"MinMoney": 1000000,
				"InterestRate": 10,
				"ElectUserNum": 3,
				"StockScale": 1600
			},
		{
				"MinMoney": 10000000,
				"InterestRate": 15,
				"ElectUserNum": 5,
				"StockScale": 2000
			}
		],
		"BlkRewardCfg": {
			"MinerMount": 3,
			"MinerHalf": 5000000,
			"ValidatorMount": 7,
			"ValidatorHalf": 5000000,
			"RewardRate": {
				"MinerOutRate": 4000,
				"ElectedMinerRate": 5000,
				"FoundationMinerRate": 1000,
				"LeaderRate": 4000,
				"ElectedValidatorsRate": 5000,
				"FoundationValidatorRate": 1000,
				"OriginElectOfflineRate": 5000,
				"BackupRewardRate": 5000
			}
		},
		"TxsRewardCfg": {
			"MinersRate": 0,
			"ValidatorsRate": 10000,
			"RewardRate": {
				"MinerOutRate": 4000,
				"ElectedMinerRate": 6000,
				"FoundationMinerRate":0,
				"LeaderRate": 4000,
				"ElectedValidatorsRate": 6000,
				"FoundationValidatorRate": 0,
				"OriginElectOfflineRate": 5000,
				"BackupRewardRate": 5000
			}
		},
		"LotteryCfg": {
			"LotteryCalc": "1",
			"LotteryInfo": [{
				"PrizeLevel": 0,
				"PrizeNum": 1,
				"PrizeMoney": 6
			}]
		},
		"InterestCfg": {
			"CalcInterval": 100,
			"PayInterval": 3600
		},
		"LeaderCfg": {
			"ParentMiningTime": 20,
			"PosOutTime": 20,
			"ReelectOutTime": 40,
			"ReelectHandleInterval": 3
		},
		"SlashCfg": {
			"SlashRate": 7500
		},
		"EleTime": {
			"MinerGen": 9,
			"MinerNetChange": 5,
			"ValidatorGen": 9,
			"ValidatorNetChange": 3,
			"VoteBeforeTime": 7
		},
		"EleInfo": {
			"ValidatorNum": 19,
			"BackValidator": 5,
			"ElectPlug": "layerd"
		},
		"ElectMinerNum": {
			"MinerNum": 21
		},
		"ElectBlackList": null,
		"ElectWhiteList": null
    },
  "config": {
					"chainID": 1,
					"byzantiumBlock": 0,
					"homesteadBlock": 0,
					"eip155Block": 0,
			"eip158Block": 0                        				             
	},
  "versionSignatures": [
    [
      181,
      8,
      246,
      28,
      118,
      103,
      127,
      70,
      144,
      31,
      187,
      28,
      71,
      14,
      164,
      113,
      133,
      96,
      141,
      160,
      117,
      234,
      127,
      5,
      254,
      240,
      146,
      127,
      39,
      247,
      161,
      150,
      75,
      243,
      248,
      192,
      32,
      110,
      149,
      242,
      151,
      195,
      226,
      167,
      74,
      223,
      135,
      250,
      233,
      174,
      109,
      239,
      101,
      177,
      155,
      129,
      68,
      92,
      218,
      222,
      45,
      207,
      165,
      112,
      0
    ]
  ],
      "difficulty":"0x100",
    "timestamp":"0x5c26f140",
		"version": "1.0.0-stable",
  
	"signatures": [	],
      "coinbase": "MAN.1111111111111111111cs",
      "leader":"MAN.CrsnQSJJfGxpb2taGhChLuyZwZJo", 
       "gasLimit": "0x2FEFD8",   
       "nonce": "0x0000000000000050",
       "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
       "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	     "extraData": "0x0000000000000000"
}
`
)

func DefaultGenesisToEthGensis(gensis1 *Genesis1, gensis *Genesis) *Genesis {
	if nil != gensis1.Config {
		gensis.Config = gensis1.Config
	}
	if gensis1.Nonce != 0 {
		gensis.Nonce = gensis1.Nonce
	}
	if gensis1.Timestamp != 0 {
		gensis.Timestamp = gensis1.Timestamp
	}
	if len(gensis1.ExtraData) != 0 {
		gensis.ExtraData = gensis1.ExtraData
	}
	if gensis1.Version != "" {
		gensis.Version = gensis1.Version
	}
	if len(gensis1.VersionSignatures) != 0 {
		gensis.VersionSignatures = gensis1.VersionSignatures
	}
	if len(gensis1.VrfValue) != 0 {
		gensis.VrfValue = gensis1.VrfValue
	}
	if len(gensis1.Signatures) != 0 {
		gensis.Signatures = gensis1.Signatures
	}
	if nil != gensis1.Difficulty {
		gensis.Difficulty = gensis1.Difficulty
	}
	if gensis1.Mixhash.Equal(common.Hash{}) == false {
		gensis.Mixhash = gensis1.Mixhash
	}
	if gensis1.Number != 0 {
		gensis.Number = gensis1.Number
	}
	if gensis1.GasUsed != 0 {
		gensis.GasUsed = gensis1.GasUsed
	}
	if gensis1.ParentHash.Equal(common.Hash{}) == false {
		gensis.ParentHash = gensis1.ParentHash
	}

	if gensis1.Leader != "" {
		gensis.Leader = base58.Base58DecodeToAddress(gensis1.Leader)
	}
	if gensis1.Coinbase != "" {
		gensis.Coinbase = base58.Base58DecodeToAddress(gensis1.Coinbase)
	}
	if gensis1.Root.Equal(common.Hash{}) == false {
		gensis.Root = gensis1.Root
	}
	if gensis1.TxHash.Equal(common.Hash{}) == false {
		gensis.TxHash = gensis1.TxHash
	}
	//nextElect
	if nil != gensis1.NextElect {
		sliceElect := make([]common.Elect, 0)
		for _, elec := range gensis1.NextElect {
			tmp := new(common.Elect)
			tmp.Account = base58.Base58DecodeToAddress(elec.Account)
			tmp.Stock = elec.Stock
			tmp.Type = elec.Type
			sliceElect = append(sliceElect, *tmp)
		}
		gensis.NextElect = sliceElect
	}

	//NetTopology
	if len(gensis1.NetTopology.NetTopologyData) != 0 {
		sliceNetTopologyData := make([]common.NetTopologyData, 0)
		for _, netTopology := range gensis1.NetTopology.NetTopologyData {
			tmp := new(common.NetTopologyData)
			tmp.Account = base58.Base58DecodeToAddress(netTopology.Account)
			tmp.Position = netTopology.Position
			sliceNetTopologyData = append(sliceNetTopologyData, *tmp)
		}
		gensis.NetTopology.NetTopologyData = sliceNetTopologyData
		gensis.NetTopology.Type = gensis1.NetTopology.Type
	}

	//Alloc
	if nil != gensis1.Alloc {
		gensis.Alloc = make(GenesisAlloc)
		for kString, vGenesisAccount := range gensis1.Alloc {
			tmpk := base58.Base58DecodeToAddress(kString)
			gensis.Alloc[tmpk] = vGenesisAccount
		}
	}

	if nil != gensis1.MState {
		if gensis.MState == nil {
			gensis.MState = new(GenesisMState)
		}
		if nil != gensis1.MState.Broadcasts {
			broadcasts := make([]common.Address, 0)
			for _, b := range *gensis1.MState.Broadcasts {
				broadcasts = append(broadcasts, base58.Base58DecodeToAddress(b))
			}
			gensis.MState.Broadcasts = &broadcasts
		}
		if nil != gensis1.MState.Foundation {
			gensis.MState.Foundation = new(common.Address)
			*gensis.MState.Foundation = base58.Base58DecodeToAddress(*gensis1.MState.Foundation)
		}
		if nil != gensis1.MState.VersionSuperAccounts {
			versionSuperAccounts := make([]common.Address, 0)
			for _, v := range *gensis1.MState.VersionSuperAccounts {
				versionSuperAccounts = append(versionSuperAccounts, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.VersionSuperAccounts = &versionSuperAccounts
		}
		if nil != gensis1.MState.BlockSuperAccounts {
			blockSuperAccounts := make([]common.Address, 0)
			for _, v := range *gensis1.MState.BlockSuperAccounts {
				blockSuperAccounts = append(blockSuperAccounts, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.BlockSuperAccounts = &blockSuperAccounts
		}
		if nil != gensis1.MState.InnerMiners {
			innerMiners := make([]common.Address, 0)
			for _, v := range *gensis1.MState.InnerMiners {
				innerMiners = append(innerMiners, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.InnerMiners = &innerMiners
		}
		if nil != gensis1.MState.ElectBlackListCfg {
			blackList := make([]common.Address, 0)
			for _, v := range *gensis1.MState.ElectBlackListCfg {
				blackList = append(blackList, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.ElectBlackListCfg = &blackList
		}
		if nil != gensis1.MState.ElectWhiteListCfg {
			whiteList := make([]common.Address, 0)
			for _, v := range *gensis1.MState.ElectWhiteListCfg {
				whiteList = append(whiteList, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.ElectBlackListCfg = &whiteList
		}
		if nil != gensis1.MState.ElectMinerNumCfg {
			gensis.MState.ElectMinerNumCfg = gensis1.MState.ElectMinerNumCfg
		}
		if nil != gensis1.MState.BlkRewardCfg {
			gensis.MState.BlkRewardCfg = gensis1.MState.BlkRewardCfg
		}
		if nil != gensis1.MState.TxsRewardCfg {
			gensis.MState.TxsRewardCfg = gensis1.MState.TxsRewardCfg
		}
		if nil != gensis1.MState.InterestCfg {
			gensis.MState.InterestCfg = gensis1.MState.InterestCfg
		}
		if nil != gensis1.MState.LotteryCfg {
			gensis.MState.LotteryCfg = gensis1.MState.LotteryCfg
		}
		if nil != gensis1.MState.SlashCfg {
			gensis.MState.SlashCfg = gensis1.MState.SlashCfg
		}
		if nil != gensis1.MState.BCICfg {
			gensis.MState.BCICfg = gensis1.MState.BCICfg
		}
		if nil != gensis1.MState.VIPCfg {
			gensis.MState.VIPCfg = gensis1.MState.VIPCfg
		}
		if nil != gensis1.MState.LeaderCfg {
			gensis.MState.LeaderCfg = gensis1.MState.LeaderCfg
		}
		if nil != gensis1.MState.EleTimeCfg {
			gensis.MState.EleTimeCfg = gensis1.MState.EleTimeCfg
		}
		if nil != gensis1.MState.EleInfoCfg {
			gensis.MState.EleInfoCfg = gensis1.MState.EleInfoCfg
		}
		//curElect
		if nil != gensis1.MState.CurElect {
			sliceElect := make([]common.Elect, 0)
			for _, elec := range *gensis1.MState.CurElect {
				tmp := new(common.Elect)
				tmp.Account = base58.Base58DecodeToAddress(elec.Account)
				tmp.Stock = elec.Stock
				tmp.Type = elec.Type
				sliceElect = append(sliceElect, *tmp)
			}
			gensis.MState.CurElect = &sliceElect
		}
	}
	return gensis
}

func GetDefaultGeneis() (*Genesis, error) {
	genesis := new(Genesis)
	defaultGenesis1 := new(Genesis1)
	err := json.Unmarshal([]byte(DefaultJson), defaultGenesis1)
	if err != nil {
		return nil, err
	}
	genesis = DefaultGenesisToEthGensis(defaultGenesis1, genesis)
	return genesis, nil

}

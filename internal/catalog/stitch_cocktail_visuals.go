package catalog

import "strings"

type CocktailVisual struct {
	Name     string
	Label    string
	Alt      string
	ImageURL string
}

var stitchCocktailVisuals = map[string]CocktailVisual{
	"aperol spritz": {
		Name:     "Aperol Spritz",
		Label:    "Citrus Bloom",
		Alt:      "Bright orange aperol spritz with an orange slice",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBahn-GjRjlVSDZYYNSNZh3gRavPAzcvfVgSbEl1k5KzHKHFjUM8YLi0-g8GNIgXYXW8LmypNlT0E-wmaxphB5JcJVI_6dfKOLEZHSIOVgmsf8-0huXc323eznk8ztsyygtTFnn50kqXFggBQN738tGMFLyX-Or_V6r7HZdBKyjm8xFqKv9d7VbWzc_H0SySKCKY-axlLDNb2TS0wYvV75_BgdPWewOBRuD-_aKIwu0w8s5s36ecSAWMUKSAYbJGUpNd3fMLvAka8g",
	},
	"beer (lager)": {
		Name:     "Beer (Lager)",
		Label:    "Light & Golden",
		Alt:      "Golden lager with condensation on glass",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuDAC0tN2QIZjHtkyyJSR_4eABlkm_CMAEpM7X2RWnnbPkN78KjO0hhI3r9jw2lELo64HJwIcUAvpRXRVHGAg3pZzmn15PpzKFgoevgYDvGNFq8thkTEc075W288ZGW-vpcWjASJUk3TXOYk52xAgSybeXoP5JgAMC_XUK8OzjEHD5-ryxUf8veLl2JyX3OznNiEHbZHxMlzhc2zJGAXzf2_Y7kNxCHNjc7oWGdP3y7LKYcmylBwEP08Fc89xJ8f7E3aKES6KdfchBc",
	},
	"bloody mary": {
		Name:     "Bloody Mary",
		Label:    "Heirloom Vine",
		Alt:      "Savory bloody mary with celery and spice",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuDQtjgKcam_JFXucpQ2hEkm-r8RuwYTWu2irErGmOd-v5j07LpZj7eUSUls7le2JRsBVIaKKWFEeFuJB-1jD2z2Dk0ILfXFgiWCcu_zEpx4dp9lvZq8ScQ_LxH0Z5v2QLPlgXQaLcw7HMPdzEeBt1NRRYde0NAAptRaJZxDyWOeBwogpl-2omhXepPfCUrdcMMtyfvp4XT43K1VJU5vGaRYfiFFwrIIqwNvQZ5mzBVtyZVYntfqJ_bZZ7V8wWGyHiluwm2ad5a9oIo",
	},
	"classic martini": {
		Name:     "Classic Martini",
		Label:    "Juniper Essence",
		Alt:      "Clear martini with green olives",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBp7ouKZpMb8V3RuFZjU4TaGR3Y7sER6Ebwd4qJw2rLVOJl0epfoMrnGIdsREUIh9iBtGE7XyN8hLuixzyj18Yf7l8P3JpogvmyzkhIkgMCvZAZgBMejKiXi6GJ_Fwe3cCFb94Eh22wJzptfwsRc2PkfMpelH6tm8z0GYrLMKl1JfMCnD31FVC0mCM17CBWHbOaAvl8VprPd7PqiyCEDzQDPCf8ehgXcPRX8tV7HjGtIXEXWwd81BlRFSKxiUEqfHdu6XzHU2s2AWY",
	},
	"cola": {
		Name:     "Cola",
		Label:    "Traditional",
		Alt:      "Cola with lime",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuD43gBNIuGCxk7olS9cCqovsGsJJFahYUF6_rYPHLq6rMYsKcNqhUbq5Arqasf9fu2b3vaTC2WupLHVjHz8f7YzXpeqXW2_EIisFrV09kAmy1JQ6i4rZGp0FN66zBn2xRErHdQBF9DjFqSMIpV3jYw_dqbAJ2U5JjY1juayti1OyW4OS-VtaN-StcrtrIUMzQJXGEo2c1w3qEWwhhsUvDJ_7_dHetXTJKsu03-Ac5HzmNO1B9Wq3uu3rToX2wn5O9cL5hoApP1k6XA",
	},
	"cosmopolitan": {
		Name:     "Cosmopolitan",
		Label:    "Berry Infusion",
		Alt:      "Vibrant pink cosmopolitan in a martini glass",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBY2_v85GVG8DjuTC-Z7tGLhrMuUi_-WJ_pY9OHAvLL4RUyNvCLD9J-EEqq5wtwvRe8RFeIazlBWhl7Y5i59Sa0hXV3z3QX4AnYeNGvIp0GeshSl2nIiCZIf911pyjV-ppkmrPn1uKNv4FKblvtNIcwzAbdVJdaN0UcQ1SuO5Dk3xExrSS5CeuZMCdH7_P1uLuHW_btUdmvfX3F4vv4jD4y0kAIZcQk6fdLISBtXUxaozKxOj4QZ8enkkwG-w-oscWR6dNNc1GN180",
	},
	"cranberry fizz": {
		Name:     "Cranberry Fizz",
		Label:    "Tart & Floral",
		Alt:      "Deep red cranberry fizz with berries",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuB4o64H-AhVrO-nYlSjgvaJFBh5jVBAnxobkCzqg7Gm-Oez8IHlsool1Z3lhGH_d1Tuh1dzjbKXNvixxFpMtE7ETHQo4Qz8XPJtJSQ_LKWq-9XNk7Ykn5T7mOghoMjHrA-AlbUDLnWcBvl7xcUDbN2np0CEA0-QttLzTrqhUZeJlpIJ-zEnUIFFKUYBW7wnxIkqxCnd4cmTw6FWYQBJhDjPe-dUzESgb_HER-U6cgIMM86vKT9165wWYhLxJJzz2cdntdc6zu4Pldc",
	},
	"cuba libre": {
		Name:     "Cuba Libre",
		Label:    "Classic Dark",
		Alt:      "Cuba Libre",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuCQnQNI47ELsSRxpNWMSBEcooPd6sIPqi9_wvmIRtYaRSXWQ9nW53PjJ_uKDVPDJozVbhFO_tNOPB0dq53ZmomL8-v9vcUWWZA25qyxATueLYm-a5pn-GKJkOkP7J6cnAA104PueKLjkv0PWnGvPfzoCYxmfH5fMjD_JRhG2NeQHC2e8zD8_CTv3BSRyXc8aDAY_QfPcVnNHfo31612YtrXz-oa49-kG3k6wmfUoYtp1L7HsP50EIeW1kkOtTrgKkY16mMZg_9L8_A",
	},
	"daiquiri": {
		Name:     "Daiquiri",
		Label:    "Lime Harvest",
		Alt:      "Fresh lime daiquiri in a coupe glass",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuB5F6MxHRkxHP9cFmB0W6cwijrzVbBAg5rwg3ZbQ0wKDozvBj8EImgetCtcCvVZNTX2RyJdT-eczgUH4pTdeaL6Txf8GyZzq5tdki9MgSqWE5OI-Tz4EP5KVSpZULH7h0P0QaAP_O2eaybsDY4qu9byh_NUXTT6UBaK1Ei6bfwInJjMxd5fRbLshGRQ6oc1wI0vY4OstJBi3ZfBm9Wb3jsOA6wpDJmbDqbM8ivZp9rROpejs0MAi2L4KIO0X7fWIrcNyxIs3GWAWWw",
	},
	"espresso martini": {
		Name:     "Espresso Martini",
		Label:    "Botanical Roast",
		Alt:      "Dark espresso martini with three coffee beans on top",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuAbSQyoLM7O3OvuUn2R3s1NsSeurTejh_su_T4_nms3AVmM0g7N6uMpC083yyE47YNGiMwDsjDBrdFvvtccg4R1v1I8DLTvj3SDYbGbwWA9oKfe66UvOqF1iZ1GvLT2a4YdP3RxwgEitZTllQvFk_rCSanumSa6U0L6YfTJ8XX2qTYq9WBrIP7sqBx5X8mfxxMfN5mEOP0LoVoNQfG8N3vsNxw_yIPXGkLtd3CB6Kf6puD-voSNRvtMbAJRpqZ6a7yz49AqxdsWVTQ",
	},
	"gin & tonic": {
		Name:     "Gin & Tonic",
		Label:    "Spirit Forward",
		Alt:      "Gin and Tonic with cucumber",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuAAliIDsgWMd3LFtx5Va1C6y0qWCWVj_M5BJJEbQwqja6mCFHp3DvEmX-Pcnt2FYrxmOifEYkWfKT1BbSgLn-VEM_Aib0y2y6NFDDaKQwsaDeuUHKU2ydsPsz4sD8alpuO5ifaof1PVzNZNCQrSsb33Q2jJ8XSLFkjnpG_Seb483YtbPhWz4Vk_7-eG27ss-s8s2ZbDkhLE6AFTtqf7IX_O_dpkYznaY-htSy4bZJ317M1PZ-pFCJpwy8IlW0kGZ0mU3bpI2tRiias",
	},
	"ginger lime fizz": {
		Name:     "Ginger Lime Fizz",
		Label:    "Sparkling",
		Alt:      "Ginger Lime Fizz",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBxJcd7aeDnGmdoPU0Gq6Huwz_f4D3trWVGdEzfZY9h3v3eRfLJNsKb7nc8uEzv0OS8hFQcP5DcIkGln8IetPQ-OjRMlxrEy_HL9V6OG5eBAxNwGF-AcsKl0e9qJTzCoEjAjyxYbD_ID8Z-RZAqjrdxseINvxoAR8yDnGe_LdmHejhU6hrAWpfltj8GZ0_fvk6G2Hnn-74LyeLQMief_x8rI2gafttA9ENSejCzQhpRITdGGozZjXn5-w-opGZYHwZvh2GRszfi048",
	},
	"lime soda": {
		Name:     "Lime Soda",
		Label:    "Botanical & Zesty",
		Alt:      "Refreshing lime soda with fresh mint sprigs",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBOgfFnhpBdxW4t4eNu6E92_ZIBz1a02g9KlZOUnZ9KP_0cAHsGOCtXZi1d4Mi87MOGUMjy49KzvlZRaE-2n6BnsXh75L_hDCNuvTxpXHzr1SfcyJZN_I6Udcns0yV7qh5Qoi1ZlvKfcCPYl_BnrLwazdkFG-9WihlwkGoPE6EDELX_86A1XxW93BkLxIIv5k-oOWC1agFH6ZNRLs6H-rgfvb2vF-bC2jyJ4ZKLA9fXA75qEQ7ZXastv8F8G4NyHqEspyD5rEVF7js",
	},
	"mai tai": {
		Name:     "Mai Tai",
		Label:    "Almond & Lime",
		Alt:      "Complex mai tai with mint and lime",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuD4AWwmjiGABdwGjVmPaErYrPTRL9j_PqhVdPF4irCfgD1PdlvxD7xRQJemky3aZauxco2mvUBU36dyPU5CyVwfQtx9jUx1jl1B7EE8M-7ZA_JAFNpgbug4P6ZQtndVS0OTyx49msf-KC1gI6H3y5JBIruFk33fGcNqn-QIpFcStNAmCQ-Zg4iT3LhIr2IUZzFOUgfUKVEdP_etWc8rnFMY0h1kGGd8x_7rvholqgXrnBPtYUCxWLbdbdeoCvxlKNPz0vzjmKZ-v6A",
	},
	"manhattan": {
		Name:     "Manhattan",
		Label:    "Oak & Cherry",
		Alt:      "Deep amber manhattan cocktail with a cherry",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuC3v4ShFt4i-UPlxRxoZZ25ZHwTDqenwFQ7ruruQqgnpnFgva0UvCEKsSGXYPmIqlmJDw6w6oSvA7WgYS2gsvcFp0qja_G-iMb7Kllk3a1q0LFRD3JA3pYozgT0vkygU6xRU_0DpclXw26jjTCaX9oVLTUQO11wnoX12ZFMW5CT6IJBOfrPD_x7zgn-oy2iGFpGHbUnCw_JYv5WY8rnSef4lOHE-JWEaJeZKCUtesFWSxreSb9ShYDRA01NM1OXE_ZwrZf0KHmabJU",
	},
	"margarita": {
		Name:     "Margarita",
		Label:    "Citrus High",
		Alt:      "Margarita with lime",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuB4TOQGSLiesfYcrtYUfGB3Vq01LxV1YKSHKnhG6udZUwnUfa0Rp2MJXIkWyNTSKcP7iMa785Q8yMHn0YpFaXZrzXC7pN63VEYwevBVMWI1qUGvx3FizfW3AZMb69gz6AB3WgwYGQAswL3-SEWMBnYBg71OZ8YX2AMM0AbLh-qhtcNdU71CWlKq2cMrV72Pg9a_dDHi-5DDPJeqFAAmcOUvMtbvNSQyCG3LA-HyqazfvunZOMAP3aus3Tlk-wcdbXBu0xQbFQxkmJc",
	},
	"mimosa": {
		Name:     "Mimosa",
		Label:    "Sunrise Nectar",
		Alt:      "Elegant mimosa in a champagne flute",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuDiH7AR1EadSgkmdoHEBiyLoIMRxX1Off3_bIyGm02C9qGKzcYhyF17BCsL8A9mNNmRqmWsDW93RcNDc7iAJks1vkQXTyM7BNwin7QNfRfw1WL_ZoBypoRz1AS3LPiFPlzosV58E6BZi-5k6crr7wB-EPYsdyjICs5FRrYDa5P9hgQWVzhQ-HWWRHdB7BXn4ijyrefvQJj7saBeUKSSfTILgGXTpVdO8eeEsixRftIjxDcXJZl1MAb29ThyoRCFc9dZS0_ED6of8us",
	},
	"mojito": {
		Name:     "Mojito",
		Label:    "Botanical",
		Alt:      "Mojito with mint",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuA6t57EfmmNtAXXKs6K3cQFNQC3C6LITBf16CoALkmC6MrkPmL6CUmDFZgLEXvTSIf9zr5HM3CcVrX2Hz2xtmvY0moOehaHrKzpe8_vD4CjDOacAbIx4XmItlkTLROw9oTdg9JZzfiUbJMG3pEJjQvmlHBRSrNDUD97HpfRNsDjOcB99YMWCh8rONRftSNohhCkifyToUCCJy0JR5tXggqnOYK0lAtsSXpZpBh9mhRTVgSH8UqsnumRcmcV2IOSwYFfzSSswWZfCRw",
	},
	"moscow mule": {
		Name:     "Moscow Mule",
		Label:    "Spicy Herb",
		Alt:      "Moscow Mule in copper mug",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBwUAFq60CdedNhRzTITfwJ141hICka_EiJ0MGCAgVZZ1JvO_c7bdCLL11ryj5J22p5PWOIDRezL6ibDkC7z_31-Zn2whDnn01lLlK4tvkxNQ2h8PA7O6tgzcctGSw8wZ2TnpGHImyhiSmznrZZexxjhl00p3C3s-dsJbVVHnBw-U4coOgqSkMl-WNXAKmC2z4Grih-E0Dx1Z-6tKV-fz4hrZkPTZ-K_jtEB8syiGnnzLfXKtRUiqiprEcUBdaRNPpykpQB08Z1t_A",
	},
	"negroni": {
		Name:     "Negroni",
		Label:    "Herbal & Intense",
		Alt:      "Vibrant red negroni with orange twist",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuAvx2IqYqoVX5IGCu_Kanf5hWLaOS8Ycz0-VAKsW5wOQfxV5lQ654Ukz0Kko2H4FSCirDubxH4z8K9X7PpVDjTf-KE9XQAnGeSzkSMz41Etlr8XoIQzTWDEm7P0I_thfV9L8Dpmf5dTmi1ULYm9tt9f6ZLcjB2GmzkO02_0LtvlFDflQ-lbTd_ATPGnsNlP5jZsRmMUso2KCsLdo9R7uAiT0-kictov1NFkOeRpLBeXU5L1FOInt0Af0b4ibpDzC6-AOhpFgfXz5nk",
	},
	"non-alcoholic beer": {
		Name:     "Non-Alcoholic Beer",
		Label:    "Hoppy & Crisp",
		Alt:      "Frothy non-alcoholic beer in a glass",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuCbudtmlTyrAglnfHaq6d3T5QybIr1nxLJJ236xVGg6O0i-HZyv8XBD9AiQnA-CFSodET3TUzNYapyAT-qo0UH0eXj7sItXIocoGpQYlfyJphsj7Ql5opnKGZl41O8yHwN6DvC6yF0QGlthf1ScDl04W1CdZWOVtEV-m12Jz8fk0486_nEYPzGtKBaH1IDAuBDR_zcumgZf_UX_M4UyxZJoQZw6H4bF5doc4x6GYLxXlTwz3-C-59XGRU_Wa7jiln45MB5dB-70zig",
	},
	"old fashioned": {
		Name:     "Old Fashioned",
		Label:    "Aromatic & Oaky",
		Alt:      "Amber old fashioned with orange peel",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuArJJh8Z_OutiYF3AZmzCXXkLxJSSYhiCgHHJTn2Exnq2Heecs4tnX3SZFb1-ja2sywpNHlgo77qWjafSJOhvrU0wPZrlWgiJDszbZBQwEJMoQvb73Y6Ct0atbkSPawKpJMLVetN5U1tXUkR2rCMSByxz_mbiRc2Y9PPiMlB4X4hAjT9EgeWObciCottZ-r_PSGhO0A2N90yuRQg8U9C2VZ_tF3Ti2nFuUI2NMoKpBCoyO0HGOUA-AmKEGBNLzdD7KGiEnhOL8pFtM",
	},
	"orange spritzer": {
		Name:     "Orange Spritzer",
		Label:    "Bright & Bubbly",
		Alt:      "Vibrant orange spritzer with citrus slices",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBdVKmc0Sk2yKyuew2foL47BBQQhrAHBkBc6Y_BVKnTypIE18yWvxr4wp_kCqd-Ps0QqNOT_601rEJJV6oV5N-ncwuESs5SIRSuYdQAQJAUZ9CExGjrwEcN_JXKxMoXade1mWGECqTJK7hJ8dxrFQLVJocwmlv4uuh0K126xJwKVJRUXoGf_fsPhCHszzRaTyL4o3L0TApCgWaa1UuYxCREjUSmSeo42pOIwkR-3ByGHe63plEL2Kc9kpBUCbYkXN6NulmjTzPXemM",
	},
	"paloma": {
		Name:     "Paloma",
		Label:    "Grapefruit Grove",
		Alt:      "Pink grapefruit paloma with a salt rim",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuA7GOQheXTtQk9C9QJANbNTctjcJScRXl9lRe2WdWhjXcc_ZsT74hekI4xRuOu_IpDtvE27fsAP_OGVf7jFCJGLG7oN0foCyEClOWRdzhxJ6_L0202zYdg_2TnCaffLzFBNcSXZsIMjWacf7N6xrBbwknfnbckpQbdI9PYTNTu-ASlsCj8LlItaAuRnbN09WGbIFZ5sSbTHxHn_HKdWhqao2id_T3aZS1mDvucaia2FLO-LhZDYDoiPj4GZgsUuvsn9ZqxxWKq022U",
	},
	"pina colada": {
		Name:     "Pina Colada",
		Label:    "Coconut Shade",
		Alt:      "Creamy pina colada with a pineapple wedge",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuC4Qj3l65tgUZ7yRJOlvAon3IEXvDVH29y8YHqUKQdH5TELrwBXt8MEnPgKNYpTL3pxMLzK6ZQgkXqs1WixW7fcLvHKdEyIZVzVMvHDqgzmg24aMAPjdmHoOrY2d6JHTTB9-S2jui0TrgwatpgzZzlZClUkAs2_cVhPw_iJFtncHkmkit8m5_a3dBS_z05LMHWxKSy2is8RKlNeCiKfATkuq2ZDQGhZb9MHsq18SBOC85QTJH31Y4r7j6j-1L3wo94OCNcYOv_7FEc",
	},
	"red wine": {
		Name:     "Red Wine",
		Label:    "Velvety & Earthy",
		Alt:      "Elegant glass of deep red wine",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuAdDdpRqIK3AHfyEpJxtaACoR_9_mGWRCb8kULtbKfKttZPUiWzHdFMG9wbrkvPYApCCziIhNrqBWUZg1EvnAr8mh0Y02hwl92JmBUiC-8nhwzl69ErQ0LXjZWkscjxbMICmeStV7ZaTBH9tN9yeOz_D2-jSomNQF-7eeg6HbaxkDd3U2frArU5hJ9eps10pw2Y6p7-Zq6Wpt685XBS3sQtd06y4b0yp16lSh9x7BzfLU5-RnWbaiFj6TQk_omUY__97RiK00W4v2o",
	},
	"rum & ginger": {
		Name:     "Rum & Ginger",
		Label:    "Warmth",
		Alt:      "Rum and Ginger",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuDa-FbmxjKu7dWpvwrOyreQ_XCPXSKmoncuxJ3Q-XE0nUOmFhCG8JG0zAHlgE3taPSocGgwoXdKB2KRY5BAIABjEr9k0dxc_atkxxmIY3-CvtKQI_rbFc6KEpgDT7zCBvjmiyL8LhuUCjle9aMLXCNmLkrrBcYHaYNrYrUvNbsj9zEBUc-EhlfIJU-6fV9le1KeeuL_OsWP8IJniHKesyFhqsuW6ExPwfFL8rRp1PnOVz807fs33xVdfC6Xs-72R06XAhW8_fCULss",
	},
	"tom collins": {
		Name:     "Tom Collins",
		Label:    "Sparkling Lemon",
		Alt:      "Refreshing tom collins with a lemon wheel",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuCvQdlkG5fT8aHGrBOxZPEv7Ml_cLRfIq2d_G2GQoIjn6w21LZdVGmbuYiArH3YUvW5pg9z_j2ct1tKNFaSF4B9X8jAiW6wNAGveXwMJuShJPlf9-FZleOJ6tKnT6YjEmH6yBtDKLXKjdIdDOT5OYAxPnwoDygHnb05kj_56JGsJGRWsxuUPiECXKalFyEynABA-fd_AEgSa-fSSEXz8lDMJHW4J66W5dBd4LDc4vZE490Bovxcu6Hra_VmiGf-FEPUTLlNBG7wwa0",
	},
	"virgin mojito": {
		Name:     "Virgin Mojito",
		Label:    "Pure Botanical",
		Alt:      "Virgin Mojito",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuDfdqrV-CbwdbpLx5x33f-H2T05keeCs51nOoUCr844xmz0QRVN9tTzUtAwC0EJKwijVlM0JADY74zFNO30Acup2QLC1Qu6BYcrWyORNLzg-3JOzwXNx5c5StsMkZcU8iM3cO1r2WJp6rws6U7hgxmjMVTysSgZyVDX4Dnb8ZjClUHoZBPsaVPciXITywsjIGOkMn5BinH9rw0MrAgGlWTUorvS0EUKP5V6qtcbscKw2GJgT29Adujv0vJSht8vuoHVokeaHaJw5tg",
	},
	"vodka soda": {
		Name:     "Vodka Soda",
		Label:    "Minimalist",
		Alt:      "Vodka Soda",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuCiUs9GIpOUQaqfBTvumb7X81OwKcSB6wVHO3gQ6UR0lWZVWn2VMIMcvd25BZJQbxcbUu-RgRpzScxZnxfea2_k2UKR0VJI5VPIPyHMjvptrFInsM74mvgk5XBx658H_bWiEHNdCJf4Vi_SOrEoqjVf8Nx2SI04IvwceUkoMI91ZHg6zgHgimOsxMjsiq2q5NCymqYYwja97s1o42NryMgAhKrHDlO5Ir66tChUNsuc9iGR2zC3RjF1TpVOMR4iOwhlJEYoyx2u5S8",
	},
	"water": {
		Name:     "Water",
		Label:    "Essence of Life",
		Alt:      "Pristine water with cucumber and herb",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuA-3gVPaJmTkWCCts5iuSE04NWWjysgrn2fxkqKNL44fhi7JOkzulSi_FcNocNawADMt6fi-8cIArGrV8sx1f5fJldvnPBo69SOQl5gTDIBZLyvUxjbcKrRXnicKinFfva_fXw_K_JZOTNH9hcellyTTE_HdgS46rC772T9QvBQKLk5CC3b4xxWTqB415aJ8g5lCQCT5N-XD5FVgOqThKiPQ_nqDJ64aA7PqaCnKOYk2XrCxCHFbTMNYK0_SQ2lCjwjj6Wh5lG6rAg",
	},
	"whiskey sour": {
		Name:     "Whiskey Sour",
		Label:    "Silky & Piquant",
		Alt:      "Whiskey sour with foam and cherry",
		ImageURL: "https://lh3.googleusercontent.com/aida-public/AB6AXuBtis1lGpfUXLC3z3WlPQN5c6lwoi2Ig8dY7FcB7MGHecq2Fiy14EEbjV9-GTWkS9MA3oYWY6A8cRKcfmMbsvC9W0Jdb5GX5Gxk7jChqgG_H8RgQ3rzq7HEh0VtLxXR5BHqyGzIG4-3sCdRQ1SUsAgNLZ7djqLfD-Bgm7r7ef07V6UTBeCgMdGs-Y1ncNcoquW9f__SrCR8DeQ7N9hPwdow1P4ZTG3Sbhh12HURpVWgcK3CPsHqG8Et0B4OFNL__Zty5I9GQstK7Q0",
	},
}

func normalizeCocktailName(name string) string {
	name = strings.TrimSpace(strings.ToLower(strings.ReplaceAll(name, "&amp;", "&")))
	return strings.Join(strings.Fields(name), " ")
}

func StitchCocktailVisualFor(name string) (CocktailVisual, bool) {
	visual, ok := stitchCocktailVisuals[normalizeCocktailName(name)]
	return visual, ok
}

func StitchCocktailImageFor(name string) string {
	if visual, ok := StitchCocktailVisualFor(name); ok {
		return visual.ImageURL
	}
	return ""
}

func StitchCocktailLabelFor(name string) string {
	if visual, ok := StitchCocktailVisualFor(name); ok {
		return visual.Label
	}
	return ""
}

func StitchCocktailAltFor(name string) string {
	if visual, ok := StitchCocktailVisualFor(name); ok && strings.TrimSpace(visual.Alt) != "" {
		return visual.Alt
	}
	return strings.TrimSpace(name)
}

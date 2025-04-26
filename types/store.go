package types

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Item struct {
	ItemTypeID    string
	ItemID        string
	Amount        int
	Name          string
	displayIcon   string
	streamedVideo string
}

type StoreItem struct {
	Item            Item
	BasePrice       int
	CurrencyID      string
	DiscountPercent int
	DiscountedPrice int
	IsPromoItem     bool
}

type Cost struct {
	ValorantPoints int
	KingdomCredits int
	FreeAgents     int
	Radianite      int
}

type Offer struct {
	OfferID          string
	IsDirectPurchase bool
	StartDate        string
	Cost             Cost
	Rewards          []Item
}

type ItemOffer struct {
	BundleItemOfferID string
	Offer             Offer
	DiscountPercent   int
	DiscountedCost    Cost
}

type FeaturedBundle struct {
	ID                         string
	DataAssetID                string
	CurrencyID                 string
	Items                      []StoreItem
	ItemOffers                 []ItemOffer
	TotalBaseCost              Cost
	TotalDiscountedCost        Cost
	TotalDiscountPercent       float64
	DurationRemainingInSeconds int
	WholesaleOnly              bool
	IsGiftable                 int
}

type SingleItemStoreOffer struct {
	OfferID          string
	IsDirectPurchase bool
	StartDate        string
	Cost             Cost
	Rewards          []Item
}

type SkinsPanelLayout struct {
	SingleItemOffers      []string
	SingleItemStoreOffers []SingleItemStoreOffer
}

type Accessory struct {
	OfferID          string
	IsDirectPurchase bool
	StartDate        string
	Cost             Cost
	Rewards          []Item
}

type AccessoryShop struct {
	Accessories []Accessory
}

type PlayerShop struct {
	Bundles          []FeaturedBundle
	SkinsPanelLayout SkinsPanelLayout
	AccessoryShop    AccessoryShop
	NightMarket      []ItemOffer
}

// Request Message Embed for shop items

func RequestShopEmbed(shop_type string, player PlayerInfo, entitlement EntitlementsTokenResponse, regional Regional) []discordgo.MessageSend {

	var message_list []discordgo.MessageSend

	switch shop_type {
	case "banner":
		{

			shop := RequestFeaturedBanner(player, entitlement, regional)

			message_list = make([]discordgo.MessageSend, len(shop))

			for Index, BundleItems := range shop {

				embeds := make([]*discordgo.MessageEmbed, len(BundleItems.Items))

				for ItemIndex, Item := range BundleItems.Items {

					embeds[ItemIndex] = &discordgo.MessageEmbed{
						Author: &discordgo.MessageEmbedAuthor{
							Name:    Item.Item.Name,
							IconURL: CurrencyIDToImage[Item.CurrencyID],
						},
						Type: discordgo.EmbedTypeImage,
						Image: &discordgo.MessageEmbedImage{
							URL: Item.Item.displayIcon,
						},
						Footer: &discordgo.MessageEmbedFooter{
							Text:    strconv.Itoa(Item.BasePrice),
							IconURL: CurrencyIDToImage[BundleItems.CurrencyID],
						},
					}
				}

				message_list[Index] = discordgo.MessageSend{
					Content:         "",
					Embeds:          embeds,
					TTS:             false,
					Components:      []discordgo.MessageComponent{},
					Files:           []*discordgo.File{},
					AllowedMentions: &discordgo.MessageAllowedMentions{},
				}

			}
		}

	case "rotation":
		{
			message_list = make([]discordgo.MessageSend, 1)

			shop := RequestRotationShop(player, entitlement, regional)

			embeds := make([]*discordgo.MessageEmbed, len(shop.SingleItemStoreOffers))

			for Index, ShopItems := range shop.SingleItemStoreOffers {

				embeds[Index] = &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    ShopItems.Rewards[0].Name,
						IconURL: CurrencyImages.ValorantPoints,
					},
					Type: discordgo.EmbedTypeImage,
					Image: &discordgo.MessageEmbedImage{
						URL: ShopItems.Rewards[0].displayIcon,
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    strconv.Itoa(ShopItems.Cost.ValorantPoints),
						IconURL: CurrencyImages.ValorantPoints,
					},
				}

				message_list[0] = discordgo.MessageSend{
					Content:         "",
					Embeds:          embeds,
					TTS:             false,
					Components:      []discordgo.MessageComponent{},
					Files:           []*discordgo.File{},
					AllowedMentions: &discordgo.MessageAllowedMentions{},
				}

			}

		}

	case "accessory":
		{
			message_list = make([]discordgo.MessageSend, 1)

			shop := RequestAccessoryShop(player, entitlement, regional)

			embeds := make([]*discordgo.MessageEmbed, len(shop.Accessories))

			for Index, accessory_item := range shop.Accessories {

				var item_type string

				switch accessory_item.Rewards[0].ItemTypeID {
				case "d5f120f8-ff8c-4aac-92ea-f2b5acbe9475":
					item_type = "Spray"
				case "3f296c07-64c3-494c-923b-fe692a4fa1bd":
					item_type = "Card"
				case "de7caa6b-adf7-4588-bbd1-143831e786c6":
					item_type = "Title"
				}

				final_name := accessory_item.Rewards[0].Name

				var description string = ""

				if item_type == "Title" {
					description = accessory_item.Rewards[0].Name
					final_name = accessory_item.Rewards[0].Name + " " + item_type
				}

				embeds[Index] = &discordgo.MessageEmbed{
					Description: description,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    final_name,
						IconURL: CurrencyImages.ValorantPoints,
					},
					Type: discordgo.EmbedTypeImage,
					Image: &discordgo.MessageEmbedImage{
						URL: accessory_item.Rewards[0].displayIcon,
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    strconv.Itoa(accessory_item.Cost.KingdomCredits),
						IconURL: CurrencyImages.KingdomCredits,
					},
				}

				message_list[0] = discordgo.MessageSend{
					Content:         "",
					Embeds:          embeds,
					TTS:             false,
					Components:      []discordgo.MessageComponent{},
					Files:           []*discordgo.File{},
					AllowedMentions: &discordgo.MessageAllowedMentions{},
				}

			}

		}

	case "night_market":
		{

			shop := RequestNightMarket(player, entitlement, regional)

			message_list = make([]discordgo.MessageSend, 1)

			embeds := make([]*discordgo.MessageEmbed, len(shop))

			for ItemIndex, Item := range shop {

				embeds[ItemIndex] = &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    Item.Offer.Rewards[0].Name,
						IconURL: CurrencyImages.ValorantPoints,
					},
					Type: discordgo.EmbedTypeImage,
					Image: &discordgo.MessageEmbedImage{
						URL: Item.Offer.Rewards[0].displayIcon,
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    strconv.Itoa(Item.DiscountedCost.ValorantPoints),
						IconURL: CurrencyImages.ValorantPoints,
					},
				}
			}

			message_list[0] = discordgo.MessageSend{
				Content:         "",
				Embeds:          embeds,
				TTS:             false,
				Components:      []discordgo.MessageComponent{},
				Files:           []*discordgo.File{},
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			}
		}
	}

	return message_list

}

// Uses token to get the player's storefront (Including cycling items)

func RequestStoreFront(player PlayerInfo, entitlement EntitlementsTokenResponse, regional Regional) map[string]interface{} {

	// Need to post an empty json object to return shop

	payload := strings.NewReader("{}")

	req, err := http.NewRequest("POST", "https://pd."+regional.shard+".a.pvp.net/store/v3/storefront/"+player.puuid, payload)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var store_front map[string]interface{}

	store_front, err = GetJSON(res)
	checkError(err)

	if store_front == nil {
		return map[string]interface{}{}
	}

	return store_front

}

// Request Featured Banner shop

func RequestFeaturedBanner(player PlayerInfo, entitlement EntitlementsTokenResponse, regional Regional) []FeaturedBundle {

	store_front := RequestStoreFront(player, entitlement, regional)

	featured_bundle_data := store_front["FeaturedBundle"].(map[string]interface{})
	bundles_array := featured_bundle_data["Bundles"].([]interface{})

	bundle_structs := make([]FeaturedBundle, len(bundles_array))

	for BundleIndex, Bundle := range bundles_array {

		bundle_data := Bundle.(map[string]interface{})

		item_data := bundle_data["Items"].([]interface{})

		// Organises the bundle entries

		store_items := make([]StoreItem, len(item_data))

		for Index, s_item := range item_data {

			storeItem_data := s_item.(map[string]interface{})
			item := storeItem_data["Item"].(map[string]interface{})

			req, err := http.NewRequest("GET", "https://valorant-api.com/v1/weapons/skinlevels/"+item["ItemID"].(string), nil)
			checkError(err)

			res, err := client.Do(req)
			checkError(err)

			defer res.Body.Close()

			var weapon_data map[string]interface{}

			weapon_data, err = GetJSON(res)
			checkError(err)

			skin_data := weapon_data["data"].(map[string]interface{})

			var video_stream string

			if skin_data["streamedVideo	"] == nil {
				video_stream = ""
			} else {
				video_stream = skin_data["streamedVideo	"].(string)
			}

			final_item := StoreItem{
				Item: Item{
					ItemTypeID:    item["ItemTypeID"].(string),
					ItemID:        item["ItemID"].(string),
					Amount:        int(item["Amount"].(float64)),
					Name:          skin_data["displayName"].(string),
					displayIcon:   "https://media.valorant-api.com/weaponskinlevels/" + item["ItemID"].(string) + "/displayicon.png",
					streamedVideo: video_stream,
				},
				BasePrice:       int(storeItem_data["BasePrice"].(float64)),
				CurrencyID:      storeItem_data["CurrencyID"].(string),
				DiscountPercent: int(storeItem_data["DiscountPercent"].(float64)),
				DiscountedPrice: int(storeItem_data["DiscountedPrice"].(float64)),
				IsPromoItem:     storeItem_data["IsPromoItem"].(bool),
			}

			store_items[Index] = final_item

		}

		// Organises the bundle's item offers entries

		item_offers_data := bundle_data["ItemOffers"].([]interface{})

		item_offers := make([]ItemOffer, len(item_offers_data))

		for Index, s_item := range item_offers_data {

			item_data := s_item.(map[string]interface{})
			offer := item_data["Offer"].(map[string]interface{})
			cost := offer["Cost"].(map[string]interface{})

			dc_cost := item_data["DiscountedCost"].(map[string]interface{})

			var ValorantPoints int
			var KingdomCredits int
			var FreeAgents int
			var Radianite int

			if dc_cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
				ValorantPoints = int(dc_cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
			} else {
				ValorantPoints = 0
			}

			if dc_cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
				KingdomCredits = int(dc_cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
			} else {
				KingdomCredits = 0
			}

			if dc_cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
				FreeAgents = int(dc_cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
			} else {
				FreeAgents = 0
			}

			if dc_cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
				Radianite = int(dc_cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
			} else {
				Radianite = 0
			}

			var ValorantPoints2 int
			var KingdomCredits2 int
			var FreeAgents2 int
			var Radianite2 int

			if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
				ValorantPoints2 = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
			} else {
				ValorantPoints2 = 0
			}

			if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
				KingdomCredits2 = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
			} else {
				KingdomCredits2 = 0
			}

			if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
				FreeAgents2 = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
			} else {
				FreeAgents2 = 0
			}

			if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
				Radianite2 = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
			} else {
				Radianite2 = 0
			}

			reward_data := offer["Rewards"].([]interface{})

			Rewards := make([]Item, len(reward_data))

			for Index, r_data := range reward_data {

				data := r_data.(map[string]interface{})

				Rewards[Index] = Item{
					ItemTypeID: data["ItemTypeID"].(string),
					ItemID:     data["ItemID"].(string),
					Amount:     int(data["Quantity"].(float64)),
				}

			}

			var DiscountPercent int

			if offer["DiscountPercent"] != nil {
				DiscountPercent = int(offer["DiscountPercent"].(float64))
			} else {
				DiscountPercent = 0
			}

			final_item := ItemOffer{
				BundleItemOfferID: item_data["BundleItemOfferID"].(string),
				Offer: Offer{
					OfferID:          offer["OfferID"].(string),
					IsDirectPurchase: offer["IsDirectPurchase"].(bool),
					StartDate:        offer["StartDate"].(string),
					Cost: Cost{
						ValorantPoints: ValorantPoints,
						KingdomCredits: KingdomCredits,
						FreeAgents:     FreeAgents,
						Radianite:      Radianite,
					},
					Rewards: Rewards,
				},
				DiscountPercent: DiscountPercent,
				DiscountedCost: Cost{
					ValorantPoints: ValorantPoints2,
					KingdomCredits: KingdomCredits2,
					FreeAgents:     FreeAgents2,
					Radianite:      Radianite2,
				},
			}

			item_offers[Index] = final_item

		}

		cost := bundle_data["TotalBaseCost"].(map[string]interface{})

		var ValorantPoints1 int
		var KingdomCredits1 int
		var FreeAgents1 int
		var Radianite1 int

		if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
			ValorantPoints1 = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
		} else {
			ValorantPoints1 = 0
		}

		if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
			KingdomCredits1 = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
		} else {
			KingdomCredits1 = 0
		}

		if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
			FreeAgents1 = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
		} else {
			FreeAgents1 = 0
		}

		if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
			Radianite1 = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
		} else {
			Radianite1 = 0
		}

		cost = bundle_data["TotalDiscountedCost"].(map[string]interface{})

		var ValorantPoints2 int
		var KingdomCredits2 int
		var FreeAgents2 int
		var Radianite2 int

		if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
			ValorantPoints2 = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
		} else {
			ValorantPoints2 = 0
		}

		if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
			KingdomCredits2 = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
		} else {
			KingdomCredits2 = 0
		}

		if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
			FreeAgents2 = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
		} else {
			FreeAgents2 = 0
		}

		if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
			Radianite2 = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
		} else {
			Radianite2 = 0
		}

		bundle_struct := FeaturedBundle{
			ID:          bundle_data["ID"].(string),
			DataAssetID: bundle_data["DataAssetID"].(string),
			CurrencyID:  bundle_data["CurrencyID"].(string),
			Items:       store_items,
			ItemOffers:  item_offers,
			TotalBaseCost: Cost{
				ValorantPoints: ValorantPoints1,
				KingdomCredits: KingdomCredits1,
				FreeAgents:     FreeAgents1,
				Radianite:      Radianite1,
			},
			TotalDiscountedCost: Cost{
				ValorantPoints: ValorantPoints2,
				KingdomCredits: KingdomCredits2,
				FreeAgents:     FreeAgents2,
				Radianite:      Radianite2,
			},
			TotalDiscountPercent:       bundle_data["TotalDiscountPercent"].(float64),
			DurationRemainingInSeconds: int(bundle_data["DurationRemainingInSeconds"].(float64)),
			WholesaleOnly:              bundle_data["WholesaleOnly"].(bool),
			IsGiftable:                 int(bundle_data["IsGiftable"].(float64)),
		}

		bundle_structs[BundleIndex] = bundle_struct

	}

	return bundle_structs

}

// Request Rotation shop

func RequestRotationShop(player PlayerInfo, entitlement EntitlementsTokenResponse, regional Regional) SkinsPanelLayout {

	store_front := RequestStoreFront(player, entitlement, regional)

	skin_panel_array := store_front["SkinsPanelLayout"].(map[string]interface{})

	single_offers_data := skin_panel_array["SingleItemOffers"].([]interface{})
	single_offers := make([]string, len(single_offers_data))

	for Index, OfferID := range single_offers_data {
		single_offers[Index] = OfferID.(string)
	}

	single_item_offers_data := skin_panel_array["SingleItemStoreOffers"].([]interface{})
	single_item_offers := make([]SingleItemStoreOffer, len(single_item_offers_data))

	for Index, offer := range single_item_offers_data {

		offer_data := offer.(map[string]interface{})

		cost := offer_data["Cost"].(map[string]interface{})

		var ValorantPoints1 int
		var KingdomCredits1 int
		var FreeAgents1 int
		var Radianite1 int

		if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
			ValorantPoints1 = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
		} else {
			ValorantPoints1 = 0
		}

		if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
			KingdomCredits1 = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
		} else {
			KingdomCredits1 = 0
		}

		if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
			FreeAgents1 = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
		} else {
			FreeAgents1 = 0
		}

		if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
			Radianite1 = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
		} else {
			Radianite1 = 0
		}

		reward_data := offer_data["Rewards"].([]interface{})

		Rewards := make([]Item, len(reward_data))

		req, err := http.NewRequest("GET", "https://valorant-api.com/v1/weapons/skinlevels/"+offer_data["OfferID"].(string), nil)
		checkError(err)

		res, err := client.Do(req)
		checkError(err)

		defer res.Body.Close()

		var weapon_data map[string]interface{}

		weapon_data, err = GetJSON(res)
		checkError(err)

		skin_data := weapon_data["data"].(map[string]interface{})

		var video_stream string

		if skin_data["streamedVideo	"] == nil {
			video_stream = ""
		} else {
			video_stream = skin_data["streamedVideo	"].(string)
		}

		for Index, r_data := range reward_data {

			data := r_data.(map[string]interface{})

			Rewards[Index] = Item{
				ItemTypeID:    data["ItemTypeID"].(string),
				ItemID:        data["ItemID"].(string),
				Amount:        int(data["Quantity"].(float64)),
				Name:          skin_data["displayName"].(string),
				displayIcon:   "https://media.valorant-api.com/weaponskinlevels/" + data["ItemID"].(string) + "/displayicon.png",
				streamedVideo: video_stream,
			}

		}

		single_item_offers[Index] = SingleItemStoreOffer{
			OfferID:          offer_data["OfferID"].(string),
			IsDirectPurchase: offer_data["IsDirectPurchase"].(bool),
			StartDate:        offer_data["StartDate"].(string),
			Cost: Cost{
				ValorantPoints: ValorantPoints1,
				KingdomCredits: KingdomCredits1,
				FreeAgents:     FreeAgents1,
				Radianite:      Radianite1,
			},
			Rewards: Rewards,
		}

	}

	skin_panel_struct := SkinsPanelLayout{
		SingleItemOffers:      single_offers,
		SingleItemStoreOffers: single_item_offers,
	}

	return skin_panel_struct

}

// Request Rotation shop

func RequestAccessoryShop(player PlayerInfo, entitlement EntitlementsTokenResponse, regional Regional) AccessoryShop {

	store_front := RequestStoreFront(player, entitlement, regional)

	accessory_shop_array := store_front["AccessoryStore"].(map[string]interface{})

	accessory_data := accessory_shop_array["AccessoryStoreOffers"].([]interface{})
	accessories := make([]Accessory, len(accessory_data))

	for AccessoryID, Access := range accessory_data {

		accessory := Access.(map[string]interface{})
		offer_data := accessory["Offer"].(map[string]interface{})

		cost := offer_data["Cost"].(map[string]interface{})

		var ValorantPoints1 int
		var KingdomCredits1 int
		var FreeAgents1 int
		var Radianite1 int

		if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
			ValorantPoints1 = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
		} else {
			ValorantPoints1 = 0
		}

		if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
			KingdomCredits1 = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
		} else {
			KingdomCredits1 = 0
		}

		if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
			FreeAgents1 = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
		} else {
			FreeAgents1 = 0
		}

		if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
			Radianite1 = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
		} else {
			Radianite1 = 0
		}

		accessory_reward_data := offer_data["Rewards"].([]interface{})

		Accessory_Reward := make([]Item, len(accessory_reward_data))

		for Index, r_data := range accessory_reward_data {

			data := r_data.(map[string]interface{})

			if data["ItemTypeID"] == "d5f120f8-ff8c-4aac-92ea-f2b5acbe9475" {
				// Is spray

				sprayData := SprayData(data["ItemID"].(string))

				Accessory_Reward[Index] = Item{
					ItemTypeID:  data["ItemTypeID"].(string),
					ItemID:      data["ItemID"].(string),
					Amount:      int(data["Quantity"].(float64)),
					Name:        sprayData.displayName,
					displayIcon: sprayData.animationGif,
				}

			}

			if data["ItemTypeID"] == "3f296c07-64c3-494c-923b-fe692a4fa1bd" {

				// Is Banner/Card

				sprayData := CardData(data["ItemID"].(string))

				Accessory_Reward[Index] = Item{
					ItemTypeID:  data["ItemTypeID"].(string),
					ItemID:      data["ItemID"].(string),
					Amount:      int(data["Quantity"].(float64)),
					Name:        sprayData.displayName,
					displayIcon: sprayData.displayIcon,
				}

			}

			if data["ItemTypeID"] == "de7caa6b-adf7-4588-bbd1-143831e786c6" {

				// Is Title

				titleData := TitleData(data["ItemID"].(string))

				Accessory_Reward[Index] = Item{
					ItemTypeID: data["ItemTypeID"].(string),
					ItemID:     data["ItemID"].(string),
					Amount:     int(data["Quantity"].(float64)),
					Name:       titleData.titleText,
				}

			}

		}

		accessory_struct := Accessory{
			OfferID:          offer_data["OfferID"].(string),
			IsDirectPurchase: offer_data["IsDirectPurchase"].(bool),
			StartDate:        offer_data["StartDate"].(string),
			Cost: Cost{
				ValorantPoints: ValorantPoints1,
				KingdomCredits: KingdomCredits1,
				FreeAgents:     FreeAgents1,
				Radianite:      Radianite1,
			},
			Rewards: Accessory_Reward,
		}

		accessories[AccessoryID] = accessory_struct

	}

	accessoryShopStruct := AccessoryShop{
		Accessories: accessories,
	}

	return accessoryShopStruct
}

// Request Night Market

func RequestNightMarket(player PlayerInfo, entitlement EntitlementsTokenResponse, regional Regional) []ItemOffer {

	store_front := RequestStoreFront(player, entitlement, regional)

	nightmarket_array := store_front["BonusStore"].(map[string]interface{})

	// Organises the bundle's item offers entries

	item_offers_data := nightmarket_array["BonusStoreOffers"].([]interface{})

	item_offers := make([]ItemOffer, len(item_offers_data))

	for Index, s_item := range item_offers_data {

		item_data := s_item.(map[string]interface{})
		offer := item_data["Offer"].(map[string]interface{})
		cost := offer["Cost"].(map[string]interface{})

		dc_cost := item_data["DiscountCosts"].(map[string]interface{})

		var ValorantPoints int
		var KingdomCredits int
		var FreeAgents int
		var Radianite int

		if dc_cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
			ValorantPoints = int(dc_cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
		} else {
			ValorantPoints = 0
		}

		if dc_cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
			KingdomCredits = int(dc_cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
		} else {
			KingdomCredits = 0
		}

		if dc_cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
			FreeAgents = int(dc_cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
		} else {
			FreeAgents = 0
		}

		if dc_cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
			Radianite = int(dc_cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
		} else {
			Radianite = 0
		}

		var ValorantPoints2 int
		var KingdomCredits2 int
		var FreeAgents2 int
		var Radianite2 int

		if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
			ValorantPoints2 = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
		} else {
			ValorantPoints2 = 0
		}

		if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
			KingdomCredits2 = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
		} else {
			KingdomCredits2 = 0
		}

		if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
			FreeAgents2 = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
		} else {
			FreeAgents2 = 0
		}

		if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
			Radianite2 = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
		} else {
			Radianite2 = 0
		}

		req, err := http.NewRequest("GET", "https://valorant-api.com/v1/weapons/skinlevels/"+offer["OfferID"].(string), nil)
		checkError(err)

		res, err := client.Do(req)
		checkError(err)

		defer res.Body.Close()

		var weapon_data map[string]interface{}

		weapon_data, err = GetJSON(res)
		checkError(err)

		skin_data := weapon_data["data"].(map[string]interface{})

		var video_stream string

		if skin_data["streamedVideo	"] == nil {
			video_stream = ""
		} else {
			video_stream = skin_data["streamedVideo	"].(string)
		}

		reward_data := offer["Rewards"].([]interface{})

		Rewards := make([]Item, len(reward_data))

		for Index, r_data := range reward_data {

			data := r_data.(map[string]interface{})

			Rewards[Index] = Item{
				ItemTypeID:    data["ItemTypeID"].(string),
				ItemID:        data["ItemID"].(string),
				Amount:        int(data["Quantity"].(float64)),
				Name:          skin_data["displayName"].(string),
				displayIcon:   "https://media.valorant-api.com/weaponskinlevels/" + data["ItemID"].(string) + "/displayicon.png",
				streamedVideo: video_stream,
			}

		}

		var DiscountPercent int

		if offer["DiscountPercent"] != nil {
			DiscountPercent = int(offer["DiscountPercent"].(float64))
		} else {
			DiscountPercent = 0
		}

		final_item := ItemOffer{
			BundleItemOfferID: item_data["BonusOfferID"].(string),
			Offer: Offer{
				OfferID:          offer["OfferID"].(string),
				IsDirectPurchase: offer["IsDirectPurchase"].(bool),
				StartDate:        offer["StartDate"].(string),
				Cost: Cost{
					ValorantPoints: ValorantPoints,
					KingdomCredits: KingdomCredits,
					FreeAgents:     FreeAgents,
					Radianite:      Radianite,
				},
				Rewards: Rewards,
			},
			DiscountPercent: DiscountPercent,
			DiscountedCost: Cost{
				ValorantPoints: ValorantPoints2,
				KingdomCredits: KingdomCredits2,
				FreeAgents:     FreeAgents2,
				Radianite:      Radianite2,
			},
		}

		item_offers[Index] = final_item

	}

	return item_offers

}

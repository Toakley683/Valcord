package types

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Item struct {
	ItemTypeID    string
	ItemID        string
	Amount        int
	Description   string
	Name          string
	DisplayIcon   string
	StreamedVideo string
	Color         string // Only really available on agents
	LevelItem     string
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

func RequestShopEmbed(shop_type string, player PlayerInfo, regional Regional) []discordgo.MessageSend {

	var message_list []discordgo.MessageSend

	switch shop_type {
	case "banner":
		{

			shop := RequestFeaturedBanner(player, regional)

			message_list = make([]discordgo.MessageSend, len(shop))

			for Index, BundleItems := range shop {

				embeds := make([]*discordgo.MessageEmbed, len(BundleItems.Items))

				for ItemIndex, Item := range BundleItems.Items {

					embeds[ItemIndex] = &discordgo.MessageEmbed{
						Author: &discordgo.MessageEmbedAuthor{
							Name:    Item.Item.Name,
							IconURL: CurrencyIDToImage[Item.CurrencyID],
						},
						Image: &discordgo.MessageEmbedImage{
							URL: Item.Item.DisplayIcon,
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

			shop := RequestRotationShop(player, regional)

			embeds := make([]*discordgo.MessageEmbed, len(shop.SingleItemStoreOffers))

			for Index, ShopItems := range shop.SingleItemStoreOffers {

				embeds[Index] = &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    ShopItems.Rewards[0].Name,
						IconURL: CurrencyImages.ValorantPoints,
					},
					Type: discordgo.EmbedTypeImage,
					Image: &discordgo.MessageEmbedImage{
						URL: ShopItems.Rewards[0].DisplayIcon,
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

			shop := RequestAccessoryShop(player, regional)

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
						URL: accessory_item.Rewards[0].DisplayIcon,
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

			shop := RequestNightMarket(player, regional)

			if len(shop) <= 0 {

				message_list := make([]discordgo.MessageSend, 1)

				message_list[0] = discordgo.MessageSend{
					Content:         "`No Night Market available`",
					TTS:             false,
					Components:      []discordgo.MessageComponent{},
					Files:           []*discordgo.File{},
					AllowedMentions: &discordgo.MessageAllowedMentions{},
				}

				return message_list

			}

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
						URL: Item.Offer.Rewards[0].DisplayIcon,
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

// Request Featured Banner shop

func RequestFeaturedBanner(player PlayerInfo, regional Regional) []FeaturedBundle {

	store_front := RequestStoreFront(player, regional)

	featured_bundle_data := store_front["FeaturedBundle"].(map[string]interface{})
	bundles_array := featured_bundle_data["Bundles"].([]interface{})

	bundle_structs := make([]FeaturedBundle, len(bundles_array))

	for BundleIndex, Bundle := range bundles_array {

		bundle_data := Bundle.(map[string]interface{})

		item_data := bundle_data["Items"].([]interface{})

		// Organises the bundle entries

		// Organises the bundle's item offers entries

		bundle_struct := FeaturedBundle{
			ID:                         bundle_data["ID"].(string),
			DataAssetID:                bundle_data["DataAssetID"].(string),
			CurrencyID:                 bundle_data["CurrencyID"].(string),
			Items:                      getStoreItems(item_data),
			ItemOffers:                 getItemOffers(bundle_data),
			TotalBaseCost:              NewCost(bundle_data["TotalBaseCost"].(map[string]interface{})),
			TotalDiscountedCost:        NewCost(bundle_data["TotalDiscountedCost"].(map[string]interface{})),
			TotalDiscountPercent:       bundle_data["TotalDiscountPercent"].(float64),
			DurationRemainingInSeconds: int(bundle_data["DurationRemainingInSeconds"].(float64)),
			WholesaleOnly:              bundle_data["WholesaleOnly"].(bool),
			IsGiftable:                 int(bundle_data["IsGiftable"].(float64)),
		}

		bundle_structs[BundleIndex] = bundle_struct

	}

	return bundle_structs

}

// Uses token to get the player's storefront (Including cycling items)

func RequestStoreFront(player PlayerInfo, regional Regional) map[string]interface{} {

	// Need to post an empty json object to return shop

	payload := strings.NewReader("{}")

	entitlement := GetEntitlementsToken(GetLockfile(true))

	req, err := http.NewRequest("POST", "https://pd."+regional.shard+".a.pvp.net/store/v3/storefront/"+player.puuid, payload)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
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

func GetWeaponData(ItemID string) map[string]interface{} {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/weapons/skinlevels/"+ItemID, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var weapon_data map[string]interface{}

	weapon_data, err = GetJSON(res)
	checkError(err)

	return weapon_data["data"].(map[string]interface{})

}

// Get all ItemOffers from bundle data (Async)

func getItemOffers(bundle_data map[string]interface{}) []ItemOffer {

	item_offers_data := bundle_data["ItemOffers"].([]interface{})

	var wg sync.WaitGroup

	type ChanItem struct {
		Index int
		Value ItemOffer
	}

	output := make(chan ChanItem, len(item_offers_data))
	item_offers := make([]ItemOffer, len(item_offers_data))

	for Index, s_item := range item_offers_data {

		wg.Add(1)

		go func(Index int, s_item interface{}) {

			defer wg.Done()

			item_data := s_item.(map[string]interface{})
			offer := item_data["Offer"].(map[string]interface{})

			reward_data := offer["Rewards"].([]interface{})

			Rewards := make([]Item, len(reward_data))

			for Index, r_data := range reward_data {

				data := r_data.(map[string]interface{})

				Rewards[Index] = ItemIDWTypeToStruct(data["ItemTypeID"].(string), data["ItemID"].(string), int(data["Quantity"].(float64)))

				NewLog("Banner Item:", Rewards[Index].Name, "- ID:", Rewards[Index].ItemID)

			}

			var DiscountPercent int

			if offer["DiscountPercent"] != nil {
				DiscountPercent = int(offer["DiscountPercent"].(float64))
			} else {
				DiscountPercent = 0
			}

			output <- ChanItem{
				Index: Index,
				Value: ItemOffer{
					BundleItemOfferID: item_data["BundleItemOfferID"].(string),
					Offer: Offer{
						OfferID:          offer["OfferID"].(string),
						IsDirectPurchase: offer["IsDirectPurchase"].(bool),
						StartDate:        offer["StartDate"].(string),
						Cost:             NewCost(offer["Cost"].(map[string]interface{})),
						Rewards:          Rewards,
					},
					DiscountPercent: DiscountPercent,
					DiscountedCost:  NewCost(item_data["DiscountedCost"].(map[string]interface{})),
				},
			}

		}(Index, s_item)

	}

	wg.Wait()
	close(output)

	for Info := range output {

		item_offers[Info.Index] = Info.Value

	}

	return item_offers

}

// Get all StoreItems from data (Async)

func getStoreItems(item_data []interface{}) []StoreItem {

	store_items := make([]StoreItem, len(item_data))

	var wg sync.WaitGroup

	type ChanItem struct {
		Index int
		Value StoreItem
	}

	output := make(chan ChanItem, len(item_data))

	for Index, s_item := range item_data {

		wg.Add(1)

		go func(Index int, s_item interface{}) {

			defer wg.Done()

			storeItem_data := s_item.(map[string]interface{})
			item := storeItem_data["Item"].(map[string]interface{})

			Item := ItemIDWTypeToStruct(item["ItemTypeID"].(string), item["ItemID"].(string), 1)

			NewLog("Banner Store Item:", Item.Name, "- ID:", Item.ItemID)

			output <- ChanItem{
				Index: Index,
				Value: StoreItem{
					Item:            Item,
					BasePrice:       int(storeItem_data["BasePrice"].(float64)),
					CurrencyID:      storeItem_data["CurrencyID"].(string),
					DiscountPercent: int(storeItem_data["DiscountPercent"].(float64)),
					DiscountedPrice: int(storeItem_data["DiscountedPrice"].(float64)),
					IsPromoItem:     storeItem_data["IsPromoItem"].(bool),
				},
			}

		}(Index, s_item)

	}

	wg.Wait()
	close(output)

	for Info := range output {

		store_items[Info.Index] = Info.Value

	}

	return store_items
}

// Create new Cost checking for null entries

func NewCost(cost map[string]interface{}) Cost {

	if cost == nil {

		return Cost{
			ValorantPoints: 0,
			KingdomCredits: 0,
			FreeAgents:     0,
			Radianite:      0,
		}

	}

	var ValorantPoints int
	var KingdomCredits int
	var FreeAgents int
	var Radianite int

	if cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"] != nil {
		ValorantPoints = int(cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"].(float64))
	} else {
		ValorantPoints = 0
	}

	if cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"] != nil {
		KingdomCredits = int(cost["85ca954a-41f2-ce94-9b45-8ca3dd39a00d"].(float64))
	} else {
		KingdomCredits = 0
	}

	if cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"] != nil {
		FreeAgents = int(cost["f08d4ae3-939c-4576-ab26-09ce1f23bb37"].(float64))
	} else {
		FreeAgents = 0
	}

	if cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"] != nil {
		Radianite = int(cost["e59aa87c-4cbf-517a-5983-6e81511be9b7"].(float64))
	} else {
		Radianite = 0
	}

	return Cost{
		ValorantPoints: ValorantPoints,
		KingdomCredits: KingdomCredits,
		FreeAgents:     FreeAgents,
		Radianite:      Radianite,
	}
}

// Request Rotation shop

func RequestRotationShop(player PlayerInfo, regional Regional) SkinsPanelLayout {

	store_front := RequestStoreFront(player, regional)

	if store_front["SkinsPanelLayout"] == nil {

		NewLog("No rotation shop (Try again?)")
		return SkinsPanelLayout{}

	}

	skin_panel_array := store_front["SkinsPanelLayout"].(map[string]interface{})

	single_offers_data := skin_panel_array["SingleItemOffers"].([]interface{})
	single_offers := make([]string, len(single_offers_data))

	for Index, OfferID := range single_offers_data {
		single_offers[Index] = OfferID.(string)
	}

	single_item_offers_data := skin_panel_array["SingleItemStoreOffers"].([]interface{})
	single_item_offers := make([]SingleItemStoreOffer, len(single_item_offers_data))

	type ChanItem struct {
		Index int
		Value SingleItemStoreOffer
	}

	var wg sync.WaitGroup
	output := make(chan ChanItem, len(single_item_offers_data))

	for Index, offer := range single_item_offers_data {

		wg.Add(1)

		go func(Index int, offer interface{}) {

			defer wg.Done()

			offer_data := offer.(map[string]interface{})

			reward_data := offer_data["Rewards"].([]interface{})

			Rewards := make([]Item, len(reward_data))

			for Index, r_data := range reward_data {

				data := r_data.(map[string]interface{})

				Rewards[Index] = ItemIDWTypeToStruct(data["ItemTypeID"].(string), data["ItemID"].(string), int(data["Quantity"].(float64)))

				NewLog("Banner Item:", Rewards[Index].Name, "- ID:", Rewards[Index].ItemID)

			}

			output <- ChanItem{
				Index: Index,
				Value: SingleItemStoreOffer{
					OfferID:          offer_data["OfferID"].(string),
					IsDirectPurchase: offer_data["IsDirectPurchase"].(bool),
					StartDate:        offer_data["StartDate"].(string),
					Cost:             NewCost(offer_data["Cost"].(map[string]interface{})),
					Rewards:          Rewards,
				},
			}

		}(Index, offer)

	}

	wg.Wait()
	close(output)

	for Info := range output {

		single_item_offers[Info.Index] = Info.Value

	}

	skin_panel_struct := SkinsPanelLayout{
		SingleItemOffers:      single_offers,
		SingleItemStoreOffers: single_item_offers,
	}

	return skin_panel_struct

}

// Request Rotation shop

func RequestAccessoryShop(player PlayerInfo, regional Regional) AccessoryShop {

	store_front := RequestStoreFront(player, regional)

	accessory_shop_array := store_front["AccessoryStore"].(map[string]interface{})

	accessory_data := accessory_shop_array["AccessoryStoreOffers"].([]interface{})
	accessories := make([]Accessory, len(accessory_data))

	type ChanItem struct {
		Index int
		Value Accessory
	}

	var wg sync.WaitGroup
	output := make(chan ChanItem, len(accessory_data))

	for AccessoryID, Access := range accessory_data {

		wg.Add(1)

		go func(AccessoryID int, Access interface{}) {

			defer wg.Done()

			accessory := Access.(map[string]interface{})
			offer_data := accessory["Offer"].(map[string]interface{})

			accessory_reward_data := offer_data["Rewards"].([]interface{})

			Accessory_Reward := make([]Item, len(accessory_reward_data))

			for Index, r_data := range accessory_reward_data {

				data := r_data.(map[string]interface{})

				Accessory_Reward[Index] = ItemIDWTypeToStruct(data["ItemTypeID"].(string), data["ItemID"].(string), int(data["Quantity"].(float64)))

				NewLog("Banner Item:", Accessory_Reward[Index].Name, "- ID:", Accessory_Reward[Index].ItemID)

			}

			output <- ChanItem{
				Index: AccessoryID,
				Value: Accessory{
					OfferID:          offer_data["OfferID"].(string),
					IsDirectPurchase: offer_data["IsDirectPurchase"].(bool),
					StartDate:        offer_data["StartDate"].(string),
					Cost:             NewCost(offer_data["Cost"].(map[string]interface{})),
					Rewards:          Accessory_Reward,
				},
			}

		}(AccessoryID, Access)

	}

	wg.Wait()
	close(output)

	for Info := range output {
		accessories[Info.Index] = Info.Value
	}

	accessoryShopStruct := AccessoryShop{
		Accessories: accessories,
	}

	return accessoryShopStruct
}

// Request Night Market

func RequestNightMarket(player PlayerInfo, regional Regional) []ItemOffer {

	store_front := RequestStoreFront(player, regional)

	if store_front["BonusStore"] == nil {
		return []ItemOffer{}
	}

	nightmarket_array := store_front["BonusStore"].(map[string]interface{})

	// Organises the bundle's item offers entries

	item_offers_data := nightmarket_array["BonusStoreOffers"].([]interface{})

	item_offers := make([]ItemOffer, len(item_offers_data))

	for Index, s_item := range item_offers_data {

		item_data := s_item.(map[string]interface{})
		offer := item_data["Offer"].(map[string]interface{})

		reward_data := offer["Rewards"].([]interface{})

		Rewards := make([]Item, len(reward_data))

		for Index, r_data := range reward_data {

			data := r_data.(map[string]interface{})

			Rewards[Index] = ItemIDWTypeToStruct(data["ItemTypeID"].(string), data["ItemID"].(string), int(data["Quantity"].(float64)))

			NewLog("Banner Item:", Rewards[Index].Name, "- ID:", Rewards[Index].ItemID)

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
				Cost:             NewCost(offer["Cost"].(map[string]interface{})),
				Rewards:          Rewards,
			},
			DiscountPercent: DiscountPercent,
			DiscountedCost:  NewCost(item_data["DiscountCosts"].(map[string]interface{})),
		}

		item_offers[Index] = final_item

	}

	return item_offers

}

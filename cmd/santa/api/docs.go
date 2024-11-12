// SPDX-License-Identifier: ice License 1.0

// Package api Code generated by swaggo/swag. DO NOT EDIT
package api

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "ice.io",
            "url": "https://ice.io"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/achievement-summaries/badges/users/{userId}": {
            "get": {
                "description": "Returns user's summary about badges.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Badges"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "the id of the user you need summary for",
                        "name": "userId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/badges.BadgeSummary"
                            }
                        }
                    },
                    "400": {
                        "description": "if validations fail",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "if not authorized",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "if not allowed",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "422": {
                        "description": "if syntax fails",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "504": {
                        "description": "if request times out",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/achievement-summaries/levels-and-roles/users/{userId}": {
            "get": {
                "description": "Returns user's summary about levels \u0026 roles.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Levels \u0026 Roles"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "the id of the user you need summary for",
                        "name": "userId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/levelsandroles.Summary"
                        }
                    },
                    "400": {
                        "description": "if validations fail",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "if not authorized",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "422": {
                        "description": "if syntax fails",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "504": {
                        "description": "if request times out",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/badges/{badgeType}/users/{userId}": {
            "get": {
                "description": "Returns all badges of the specific type for the user, with the progress for each of them.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Badges"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "the id of the user you need progress for",
                        "name": "userId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "enum": [
                            "level",
                            "coin",
                            "social"
                        ],
                        "type": "string",
                        "description": "the type of the badges",
                        "name": "badgeType",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/badges.Badge"
                            }
                        }
                    },
                    "400": {
                        "description": "if validations fail",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "if not authorized",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "if not allowed",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "422": {
                        "description": "if syntax fails",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "504": {
                        "description": "if request times out",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/tasks/x/users/{userId}": {
            "get": {
                "description": "Returns all the tasks and provided user's progress for each of them.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tasks"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "the id of the user you need progress for",
                        "name": "userId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "enum": [
                            "pending",
                            "completed"
                        ],
                        "type": "string",
                        "description": "pending/completed status filter",
                        "name": "status",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/tasks.Task"
                            }
                        }
                    },
                    "400": {
                        "description": "if validations fail",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "if not authorized",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "if not allowed",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "422": {
                        "description": "if syntax fails",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    },
                    "504": {
                        "description": "if request times out",
                        "schema": {
                            "$ref": "#/definitions/server.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "badges.AchievingRange": {
            "type": "object",
            "properties": {
                "fromInclusive": {
                    "type": "integer"
                },
                "toInclusive": {
                    "type": "integer"
                }
            }
        },
        "badges.Badge": {
            "type": "object",
            "properties": {
                "achieved": {
                    "type": "boolean"
                },
                "achievingRange": {
                    "$ref": "#/definitions/badges.AchievingRange"
                },
                "name": {
                    "type": "string"
                },
                "percentageOfUsersInProgress": {
                    "type": "number"
                },
                "type": {
                    "$ref": "#/definitions/badges.GroupType"
                }
            }
        },
        "badges.BadgeSummary": {
            "type": "object",
            "properties": {
                "index": {
                    "type": "integer"
                },
                "lastIndex": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "type": {
                    "$ref": "#/definitions/badges.GroupType"
                }
            }
        },
        "badges.GroupType": {
            "type": "string",
            "enum": [
                "level",
                "coin",
                "social"
            ],
            "x-enum-varnames": [
                "LevelGroupType",
                "CoinGroupType",
                "SocialGroupType"
            ]
        },
        "levelsandroles.Role": {
            "type": "object",
            "properties": {
                "enabled": {
                    "type": "boolean",
                    "example": true
                },
                "type": {
                    "allOf": [
                        {
                            "$ref": "#/definitions/levelsandroles.RoleType"
                        }
                    ],
                    "example": "snowman"
                }
            }
        },
        "levelsandroles.RoleType": {
            "type": "string",
            "enum": [
                "ambassador"
            ],
            "x-enum-varnames": [
                "AmbassadorRoleType"
            ]
        },
        "levelsandroles.Summary": {
            "type": "object",
            "properties": {
                "level": {
                    "type": "integer",
                    "example": 11
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/levelsandroles.Role"
                    }
                }
            }
        },
        "server.ErrorResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string",
                    "example": "SOMETHING_NOT_FOUND"
                },
                "data": {
                    "type": "object",
                    "additionalProperties": {}
                },
                "error": {
                    "type": "string",
                    "example": "something is missing"
                }
            }
        },
        "tasks.Data": {
            "type": "object",
            "properties": {
                "requiredQuantity": {
                    "type": "integer",
                    "example": 3
                },
                "telegramUserHandle": {
                    "type": "string",
                    "example": "jdoe1"
                },
                "twitterUserHandle": {
                    "type": "string",
                    "example": "jdoe2"
                },
                "verificationCode": {
                    "type": "string",
                    "example": "ABC"
                }
            }
        },
        "tasks.Metadata": {
            "type": "object",
            "properties": {
                "errorDescription": {
                    "type": "string",
                    "example": "Error description"
                },
                "iconUrl": {
                    "type": "string",
                    "example": "https://app.ice.com/web/invite.svg"
                },
                "longDescription": {
                    "type": "string",
                    "example": "Long description"
                },
                "shortDescription": {
                    "type": "string",
                    "example": "Short description"
                },
                "taskUrl": {
                    "type": "string",
                    "example": "https://x.com/ice_blockchain"
                },
                "title": {
                    "type": "string",
                    "example": "Claim username"
                }
            }
        },
        "tasks.Task": {
            "type": "object",
            "properties": {
                "completed": {
                    "type": "boolean",
                    "example": false
                },
                "data": {
                    "$ref": "#/definitions/tasks.Data"
                },
                "metadata": {
                    "$ref": "#/definitions/tasks.Metadata"
                },
                "prize": {
                    "type": "number",
                    "example": 200
                },
                "type": {
                    "allOf": [
                        {
                            "$ref": "#/definitions/tasks.Type"
                        }
                    ],
                    "example": "claim_username"
                }
            }
        },
        "tasks.Type": {
            "type": "string",
            "enum": [
                "claim_username",
                "start_mining",
                "upload_profile_picture",
                "follow_us_on_twitter",
                "join_twitter",
                "join_telegram",
                "invite_friends",
                "join_twitter",
                "join_telegram",
                "join_reddit_ion",
                "join_instagram_ion",
                "join_twitter_ion",
                "join_telegram_ion",
                "join_youtube",
                "join_twitter_multiversx",
                "join_twitter_xportal",
                "join_telegram_multiversx",
                "join_bullish_cmc",
                "join_ion_cmc",
                "join_watchlist_cmc",
                "join_portfolio_coingecko",
                "join_holdcoin",
                "join_human",
                "join_hipo",
                "join_freedogs",
                "join_athene",
                "join_kolo",
                "join_ducks",
                "join_cmc_ton",
                "join_cmc_sol",
                "join_cmc_bnb",
                "join_cmc_eth",
                "join_cmc_btc",
                "join_bearfi",
                "join_boinkers",
                "join_dejendog",
                "join_catgoldminer",
                "watch_video_with_code_confirmation_1",
                "invite_friends_5",
                "invite_friends_10",
                "invite_friends_25",
                "invite_friends_50",
                "invite_friends_100",
                "invite_friends_200",
                "signup_callfluent",
                "signup_sauces",
                "signup_sealsend",
                "signup_sunwaves",
                "signup_doctorx",
                "signup_tokero",
                "claim_badge_l1",
                "claim_badge_l2",
                "claim_badge_l3",
                "claim_badge_l4",
                "claim_badge_l5",
                "claim_badge_l6",
                "claim_badge_c1",
                "claim_badge_c2",
                "claim_badge_c3",
                "claim_badge_c4",
                "claim_badge_c5",
                "claim_badge_c6",
                "claim_badge_c7",
                "claim_badge_c8",
                "claim_badge_c9",
                "claim_badge_c10",
                "claim_badge_s1",
                "claim_badge_s2",
                "claim_badge_s3",
                "claim_badge_s4",
                "claim_badge_s5",
                "claim_badge_s6",
                "claim_badge_s7",
                "claim_badge_s8",
                "claim_badge_s9",
                "claim_badge_s10",
                "claim_level_1",
                "claim_level_2",
                "claim_level_3",
                "claim_level_4",
                "claim_level_5",
                "claim_level_6",
                "claim_level_7",
                "claim_level_8",
                "claim_level_9",
                "claim_level_10",
                "claim_level_11",
                "claim_level_12",
                "claim_level_13",
                "claim_level_14",
                "claim_level_15",
                "claim_level_16",
                "claim_level_17",
                "claim_level_18",
                "claim_level_19",
                "claim_level_20",
                "claim_level_21",
                "mining_streak_7",
                "mining_streak_14",
                "mining_streak_30"
            ],
            "x-enum-varnames": [
                "ClaimUsernameType",
                "StartMiningType",
                "UploadProfilePictureType",
                "FollowUsOnTwitterType",
                "JoinTwitterType",
                "JoinTelegramType",
                "InviteFriendsType",
                "JoinTwitterUsType",
                "JoinTelegramUsType",
                "JoinRedditIONType",
                "JoinInstagramIONType",
                "JoinTwitterIONType",
                "JoinTelegramIONType",
                "JoinYoutubeType",
                "JoinTwitterMultiversxType",
                "JoinTwitterXPortalType",
                "JoinTelegramMultiversxType",
                "JoinBullishCMCType",
                "JoinIONCMCType",
                "JoinWatchListCMCType",
                "JoinPortfolioCoinGeckoType",
                "JoinHoldCoinType",
                "JoinHumanType",
                "JoinHipoType",
                "JoinFreedogsType",
                "JoinAtheneType",
                "JoinKoloType",
                "JoinDucksType",
                "JoinCMCTONType",
                "JoinCMCSSOLType",
                "JoinCMCBNBType",
                "JoinCMCETHType",
                "JoinCMCBTCType",
                "JoinBearfiType",
                "JoinBoinkersType",
                "JoinDejenDogType",
                "JoinCatGoldMinerType",
                "WatchVideoWithCodeConfirmation1Type",
                "InviteFriends5Type",
                "InviteFriends10Type",
                "InviteFriends25Type",
                "InviteFriends50Type",
                "InviteFriends100Type",
                "InviteFriends200Type",
                "SignUpCallfluentType",
                "SignUpSaucesType",
                "SignUpSealsendType",
                "SignUpSunwavesType",
                "SignUpDoctorxType",
                "SignUpTokeroType",
                "ClaimLevelBadge1Type",
                "ClaimLevelBadge2Type",
                "ClaimLevelBadge3Type",
                "ClaimLevelBadge4Type",
                "ClaimLevelBadge5Type",
                "ClaimLevelBadge6Type",
                "ClaimCoinBadge1Type",
                "ClaimCoinBadge2Type",
                "ClaimCoinBadge3Type",
                "ClaimCoinBadge4Type",
                "ClaimCoinBadge5Type",
                "ClaimCoinBadge6Type",
                "ClaimCoinBadge7Type",
                "ClaimCoinBadge8Type",
                "ClaimCoinBadge9Type",
                "ClaimCoinBadge10Type",
                "ClaimSocialBadge1Type",
                "ClaimSocialBadge2Type",
                "ClaimSocialBadge3Type",
                "ClaimSocialBadge4Type",
                "ClaimSocialBadge5Type",
                "ClaimSocialBadge6Type",
                "ClaimSocialBadge7Type",
                "ClaimSocialBadge8Type",
                "ClaimSocialBadge9Type",
                "ClaimSocialBadge10Type",
                "ClaimLevel1Type",
                "ClaimLevel2Type",
                "ClaimLevel3Type",
                "ClaimLevel4Type",
                "ClaimLevel5Type",
                "ClaimLevel6Type",
                "ClaimLevel7Type",
                "ClaimLevel8Type",
                "ClaimLevel9Type",
                "ClaimLevel10Type",
                "ClaimLevel11Type",
                "ClaimLevel12Type",
                "ClaimLevel13Type",
                "ClaimLevel14Type",
                "ClaimLevel15Type",
                "ClaimLevel16Type",
                "ClaimLevel17Type",
                "ClaimLevel18Type",
                "ClaimLevel19Type",
                "ClaimLevel20Type",
                "ClaimLevel21Type",
                "MiningStreak7Type",
                "MiningStreak14Type",
                "MiningStreak30Type"
            ]
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "latest",
	Host:             "",
	BasePath:         "/v1r",
	Schemes:          []string{"https"},
	Title:            "Achievements API",
	Description:      "API that handles everything related to read-only operations for user's achievements and gamification progress.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

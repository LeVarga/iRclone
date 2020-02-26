//
//  Providers.swift
//  iRclone
//
//  Created by Levente Varga on 1/3/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

struct Providers: Codable {
    let providers: [Provider]
}

struct Provider: Codable {
    let name: String
    let description: String
    let prefix: String
    let options: [ProviderOptions]?
    
    enum CodingKeys: String, CodingKey {
        case name = "Name"
        case description = "Description"
        case prefix = "Prefix"
        case options = "Options"
    }
}

struct ProviderOptions: Codable {
    let name, help, provider: String
    //let optionDefault: String?
    let value: String?
    let shortOpt: String
    let hide: Int
    let optionRequired, isPassword, noPrefix, advanced: Bool
    let defaultStr, valueStr: String
    let type: String
    let examples: [Example]?
    
    enum CodingKeys: String, CodingKey {
        case name = "Name"
        case help = "Help"
        case provider = "Provider"
        //case optionDefault = "Default"
        case value = "Value"
        case shortOpt = "ShortOpt"
        case hide = "Hide"
        case optionRequired = "Required"
        case isPassword = "IsPassword"
        case noPrefix = "NoPrefix"
        case advanced = "Advanced"
        case defaultStr = "DefaultStr"
        case valueStr = "ValueStr"
        case type = "Type"
        case examples = "Examples"
    }
}

struct Example: Codable {
    let value: String?
    let help, provider: String?
    
    enum CodingKeys: String, CodingKey {
        case value = "Value"
        case help = "Help"
        case provider = "Provider"
    }
}

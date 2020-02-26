//
//  RemoteFile.swift
//  iRclone
//
//  Created by Levente Varga on 2/17/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation

struct RemoteFile : Codable, Equatable, File {
    let path: String
    let name: String
    let size: Int64
    let mimeType: String
    let modTime: String
    let isDir: Bool
    
    enum CodingKeys: String, CodingKey {
        case path = "Path"
        case name = "Name"
        case size = "Size"
        case mimeType = "MimeType"
        case modTime = "ModTime"
        case isDir = "IsDir"
    }
}

extension RemoteFile {
    func delete(remote: String?, completion: @escaping (NSError?) -> Void) {
        if let remote = remote {
            let url = "operations/" + (isDir ? "purge" : "deletefile") + "?fs=\(remote)&remote=\(path)"
            Rclone.request(queryString: url, jsonData: nil, timeout: 30, decodeAs: Empty.self) { (_, error) in
                completion(error as NSError?)
            }
        }
    }
}

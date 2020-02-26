//
//  RemoteFiles.swift
//  iRclone
//
//  Created by Levente Varga on 1/22/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation
import Rclone

struct RemoteFiles: Codable {
    var list: [RemoteFile]
}

extension RemoteFiles {
    static func load(remote: String, path: URL?, completion: @escaping (Error?, [RemoteFile]) -> Void) {
        Rclone.request(queryString: "operations/list?fs=\(remote)&remote=\(path?.path ?? "")", jsonData: nil, timeout: 15, decodeAs: RemoteFiles.self) { (decoded, error) in
            if let contents = decoded?.list {
                completion(nil, contents)
            } else {
                completion(error as NSError?, [])
            }
        }
    }
    
    static func mkdir(remote: String, path: URL, completion: @escaping (Error?) -> Void) {
        Rclone.request(queryString: "operations/mkdir?fs=\(remote)&remote=\(path.path)", jsonData: nil, timeout: 5, decodeAs: Empty.self) { (_, error) in
            completion(error)
        }
    }
}

//
//  Remotes.swift
//  iRclone
//
//  Created by Levente Varga on 1/23/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation
import Rclone

class Remotes: Collection {
    
    typealias Element = Remote
    typealias Index = Int

    private var remotes: [Remote] = []
    
    var startIndex: Index { return remotes.startIndex }
    var endIndex: Index { return remotes.endIndex }

    subscript(index: Index) -> Remote {
        get { return remotes[index] }
    }

    func index(after i: Index) -> Index { return remotes.index(after: i) }
    
    init() { return }
    
    func load(completion: @escaping (NSError?) -> Void) {
        Rclone.request(queryString: "config/dump", jsonData: nil, timeout: 30, decodeAs: RemoteList.self) { (list, error) in
            if let remotes = list?.remotes {
                self.remotes = remotes
                completion(nil)
            } else {
                completion(error as NSError?)
            }
        }
    }
    
    func delete(remote: Remote, completion: @escaping (NSError?) -> Void) {
        Rclone.request(queryString: "config/delete?name=\(remote.name!)", jsonData: nil, timeout: 30, decodeAs: Empty.self) { (_, error) in
            completion(error as NSError?)
        }
    }
}

struct Remote: Decodable {
    var name: String?
    var type: String
    
    enum CodingKeys: String, CodingKey {
        case type = "type"
    }
}

struct RemoteList: Decodable {

    var remotes: [Remote]
    
    struct CodingKeys: CodingKey {
        var stringValue: String
        var intValue: Int?
        
        init?(stringValue: String) {
            self.stringValue = stringValue
        }
        
        init?(intValue: Int) {
            self.intValue = intValue
            self.stringValue = String(intValue)
        }
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)

        var remotes = [Remote]()
        for key in container.allKeys {
            if let remote = try? container.decode(Remote.self, forKey: key) {
                var rm = remote
                rm.name = key.stringValue
                remotes.append(rm)
            } else {
            }
        }
        self.remotes = remotes.sorted(by: { $0.name! < $1.name! })
    }
}


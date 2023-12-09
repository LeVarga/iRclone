//
//  LocalFile.swift
//  iRclone
//
//  Created by Levente Varga on 2/17/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation

struct LocalFile : File {
    func delete(remote: String?, completion: @escaping (NSError?) -> Void) {
        do {
            try FileManager.default.removeItem(at: URL(fileURLWithPath: self.path, isDirectory: self.isDir))
        } catch let error as NSError {
            completion(error)
        }
        completion(nil)
    }
    
    func rename(remote: String?, newName: String, completion: @escaping (NSError?) -> Void) {
        do {
            let f = URL(fileURLWithPath: self.path, isDirectory: self.isDir)
            try FileManager.default.moveItem(at: f, to: f.deletingLastPathComponent().appendingPathComponent(newName, isDirectory: self.isDir))
        } catch let error as NSError {
            completion(error)
        }
        completion(nil)
    }
    
    var isDir: Bool
    let path: String
    let name: String
    let size: Int64
    
    init(file: URL) {
        self.isDir = file.hasDirectoryPath
        self.name = file.lastPathComponent
        self.path = file.path
        if !isDir {
            let attributes = try? file.resourceValues(forKeys:[.fileSizeKey])
            self.size = Int64(attributes?.fileSize ?? -1)
        } else {
            self.size = -1
        }
        
    }
}

//
//  File.swift
//  iRclone
//
//  Created by Levente Varga on 2/17/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation

protocol File {
    var path: String { get }
    var name: String { get }
    var isDir: Bool { get }
    var size: Int64 { get }
    func delete(remote: String?, completion: @escaping (NSError?) -> Void)
}

extension File {
    var type: fileTypes? {
        get {
            if isDir {
                return .directory
            }
            let fileExtension = path.components(separatedBy: ".").last?.lowercased()
            if ["gif", "jpg", "png"].contains(fileExtension) {
                return .image
            } else if ["mp4", "mkv", "mov", "avi", "webm", "flv"].contains(fileExtension) {
                return .video
            } else if ["mp3", "wav", "aac", "flac"].contains(fileExtension) {
                return .audio
            } else if ["pdf", "doc", "docx", "xls", "xlsx", "rtf", "ppt", "pptx", "key", "pages", "numbers"].contains(fileExtension) {
                return .document
            }
            return nil
        }
    }
}

enum fileTypes {
    case video
    case image
    case audio
    case document
    case directory
}


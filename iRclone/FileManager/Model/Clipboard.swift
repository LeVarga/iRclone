//
//  Clipboard.swift
//  iRclone
//
//  Created by Levente Varga on 2/7/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation

var clipboard: Clipboard?

struct Clipboard {
    var files: [File]
    var move: Bool
    var srcRemoteFs: String?
    
    init(files: [File], move: Bool, srcRemoteFs: String?) {
        self.files = files
        self.move = move
        self.srcRemoteFs = srcRemoteFs
    }
    
    func paste(dstPath: String?, dstRemoteFs: String?) -> [Error]? {
        var errors: [Error] = []
        let fm = FileManager()
        for file in files {
            if let localFile = file as? LocalFile, dstRemoteFs == nil {
                do {
                    if move {
                        try fm.moveItem(atPath: localFile.path, toPath: dstPath! + "/" + localFile.name)
                    } else {
                        try fm.copyItem(atPath: localFile.path, toPath: dstPath! + "/" + localFile.name)
                    }
                } catch let error {
                    errors.append(error)
                }
            } else {
                Transfers.new(dstPath: dstPath, dstRemoteFs: dstRemoteFs, file: file, operation: move ? .move : .copy, srcRemoteFs: srcRemoteFs)
            }
        }
        return nil
    }
}

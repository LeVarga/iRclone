//
//  Rclone.swift
//  iRclone
//
//  Created by Levente Varga on 2/17/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation
import Rclone

class Rclone {
    private static var rcPort: UInt16 = 48725 {
        didSet {
            var err: NSError?
            RcloneStopRC(&err)
            if (err != nil) {
                print(err!.localizedDescription)
                return
            }
            RcloneStartRC("\(rcHost):\(rcPort)", &err)
            if (err != nil) {
                print(err!.localizedDescription)
            }
        }
    }
    private static let socket = tmpDirURL.appendingPathComponent("rcsocket", isDirectory: false)
    private static let rcHost = "[::1]"
    public static var rcUrl: URL {
        get {
            return URL(string: "http://\(rcHost):\(rcPort)")!
        }
    }
    public static var authState: String {
        get {
            return RcloneGetAuthState()
        }
    }
    public static var configPath: URL = documentsUrl.appendingPathComponent("config", isDirectory: true).appendingPathComponent("rclone.conf", isDirectory: false) {
        didSet {
            setup()
        }
    }
      
    public static func request<T>(queryString: String, jsonData: Data?, timeout: TimeInterval, decodeAs: T.Type, completion: @escaping (T?, Error?) -> Void) where T : Decodable {
        if let url = URL(string: queryString.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)!, relativeTo: rcUrl) {
            var request = URLRequest(url: url)
            if jsonData != nil {
                request.setValue("application/json", forHTTPHeaderField: "Content-Type")
                request.httpBody = jsonData
            }
            request.httpMethod = "POST"
            request.timeoutInterval = timeout
            
            let task = URLSession.shared.dataTask(with: request) { (data, response, error) in
                let statusCode = (response as? HTTPURLResponse)?.statusCode
                let decoder = JSONDecoder()
                var errorToReturn: Error? = error
                var decodedData: T? = nil
                if let data = data {
                    do {
                        if statusCode == 200 {
                            decodedData = try decoder.decode(T.self, from: data)
                        } else {
                            errorToReturn = NSError(domain:"", code: statusCode ?? 0, userInfo:[NSLocalizedDescriptionKey : try decoder.decode(ErrorResponse.self, from: data).error])
                        }
                    } catch {
                        errorToReturn = error
                    }
                }
                DispatchQueue.main.async {
                    completion(decodedData, errorToReturn)
                }
            }
            task.resume()
        } else {
            completion(nil, NSError(domain:"", code:0, userInfo:[NSLocalizedDescriptionKey : "Invalid URL"]))
        }
    }
    
    public static func setup() {
        if !FileManager.default.fileExists(atPath: configPath.path) {
            if !FileManager.default.fileExists(atPath: configPath.deletingLastPathComponent().path) {
                try! FileManager.default.createDirectory(at: configPath.deletingLastPathComponent(), withIntermediateDirectories: true, attributes: nil)
            }
            print("Config file doesn't exist, creating empty rclone.conf")
            FileManager.default.createFile(atPath: configPath.path, contents: nil, attributes: nil)
        }
        RcloneSetConfigPath(configPath.path, nil)
    }
      
    ///Start the remote control server
    static func start() {
        var err: NSError?
        RcloneStartRC("\(rcHost):\(rcPort)", &err)
        if (err != nil) {
            print(err!.localizedDescription)
        }
    }
    
    static func stop() {
        var err: NSError?
        RcloneStopRC(&err)
        if err != nil {
            print(err?.localizedDescription ?? "Unspecified error stopping server")
        }
    }
    
    struct ErrorResponse: Decodable {
        let error: String
    }
    
    struct Size: Decodable {
        let bytes: Int64
        let count: Int
    }
}

struct Empty: Decodable {
}

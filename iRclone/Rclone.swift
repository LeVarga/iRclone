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
    public static var rcPort: Int16 = 5572 {
        didSet {
            RcloneStopRC()
            RcloneStartRC(nil)
        }
    }
    private static let rcHost = "localhost"
    private static var rcUrl: URL {
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
                DispatchQueue.main.async {
                    let statusCode = (response as? HTTPURLResponse)?.statusCode
                    let decoder = JSONDecoder()
                    if error == nil && statusCode == 200, let data = data {
                        do {
                            let decodedData = try decoder.decode(T.self, from: data)
                            completion(decodedData, error)
                        } catch let error {
                            completion(nil, error)
                        }
                    } else if error == nil, let data = data {
                        do {
                            let rcloneErrorResponse = try decoder.decode(ErrorResponse.self, from: data)
                            completion(nil, NSError(domain:"", code: statusCode ?? 0, userInfo:[NSLocalizedDescriptionKey : rcloneErrorResponse.error]))
                        } catch let error {
                            completion(nil, error)
                        }
                    } else if error == nil {
                        completion(nil, NSError(domain:"", code: statusCode ?? 0, userInfo:[NSLocalizedDescriptionKey : "No data received, status code \(statusCode ?? 0)"]))
                    } else {
                        completion(nil, error)
                    }
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
        RcloneSetConfigPath(configPath.path)
    }
    
    static func start() {
        RcloneStartRC(nil)
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

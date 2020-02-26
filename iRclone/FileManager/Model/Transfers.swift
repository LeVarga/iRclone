//
//  Transfers.swift
//  iRclone
//
//  Created by Levente Varga on 1/6/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import Foundation

struct Transfers {
    private static var allJobs = [Job]()
    
    static var inProgress: [Job] {
        get {
            return allJobs.filter { !($0.status?.finished ?? false) }
        }
    }
    static var completed: [Job] {
        get {
            return allJobs.filter { $0.status?.success ?? false }
        }
    }
    static var failed: [Job] {
        get {
            return allJobs.filter { ($0.status?.finished ?? false) && !($0.status?.success ?? false) }
        }
    }
        
    static func new(dstPath: String?, dstRemoteFs: String?, file: File, operation: Operation, srcRemoteFs: String?) {
        print(operation.rawValue + " from " + (srcRemoteFs ?? "") + file.path)
        print(" to " + (dstRemoteFs ?? "") + (dstPath ?? "") + "/\(file.name)")
        var queryURL: String
        if !file.isDir { //is not directory
            queryURL = "operations/\(operation.rawValue)file?srcFs=\(srcRemoteFs ?? "/")&srcRemote=\(file.path)&dstFs=\(dstRemoteFs ?? "/")&dstRemote=\(dstPath ?? "")/\(file.name)&_async=true"
        } else { //is directory
            queryURL = "sync/\(operation.rawValue)?srcFs=\((srcRemoteFs ?? "/") + file.path)&dstFs=\((dstRemoteFs ?? "/") + (dstPath ?? ""))/\(file.name)&_async=true"
        }
        Rclone.request(queryString: "operations/size?fs=\(srcRemoteFs ?? "")\(file.path)", jsonData: nil, timeout: 5, decodeAs: Rclone.Size.self) { (size, error) in
            Rclone.request(queryString: queryURL, jsonData: nil, timeout: 5, decodeAs: Job.JobStarted.self) { (job, error) in
                print(job?.jobid)
                if let error = error {
                    print(error.localizedDescription)
                } else {
                    allJobs.append(Job(jobid: job!.jobid, name: file.name, size: size?.bytes ?? file.size, numTransfers: size?.count))
                }
            }
        }
    }
    
    static func remove(id: Int) {
        allJobs = allJobs.filter({ (job) -> Bool in
            if job.jobid == id {
                if !(job.status?.finished ?? false) {
                    job.stop()
                }
                return false
            }
            return true
        })
    }
    
    ///Refreshes status of all jobs in-progress
    static func refresh() {
        inProgress.forEach { (job) in
            job.refresh()
        }
    }
    
    enum Operation: String {
        case move
        case copy
    }
}

class Job {
    var jobid: Int
    var name: String
    var status: JobStatus?
    var stats: JobStats?
    var size: Int64
    var numTransfers: Int?
    
    var percentDone: Float? {
        get {
            if let stats = stats {
                return Float(stats.bytes) / Float(size)
            }
            return nil
        }
    }
    
    init(jobid: Int, name: String, size: Int64, numTransfers: Int?) {
        self.jobid = jobid
        self.name = name
        self.numTransfers = numTransfers
        self.size = size
    }
    
    ///Reloads job status and stats
    func refresh() {
        if !(status?.finished ?? false) {
            Rclone.request(queryString: "job/status?jobid=\(jobid)", jsonData: nil, timeout: 1, decodeAs: JobStatus.self) { (status, error) in
                if error == nil, status != nil {
                    self.status = status
                }
            }
            Rclone.request(queryString: "core/stats?group=job/\(jobid)", jsonData: nil, timeout: 1, decodeAs: JobStats.self) { (stats, error) in
                if error == nil, stats != nil {
                   self.stats = stats
                }
            }
        }
    }
    
    ///Cancels the job
    func stop() {
        Rclone.request(queryString: "job/stop?jobid=\(jobid)", jsonData: nil, timeout: 1, decodeAs: Empty.self) { (_, error) in
        }
    }
    
    struct JobStatus: Decodable {
        let finished: Bool
        let error: String
        let success: Bool
    }
    
    struct FileTransfers: Decodable {
        let size: Int64
    }
    
    struct JobStats: Decodable {
        let speed: Double
        let bytes: Int64
        let transferring: [FileTransfers]?
    }
    
    struct JobStarted: Decodable {
        let jobid: Int
    }
}

//
//  TransfersTableViewController.swift
//  iRclone
//
//  Created by Levente Varga on 2/11/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit

class TransfersTableViewController: UITableViewController {
    var timer = Timer()
    
    override func viewDidLoad() {
        super.viewDidLoad()
        tableView.rowHeight = 84
    }
    
    override func viewWillAppear(_ animated: Bool) {
        timer = Timer.scheduledTimer(timeInterval: 1, target: self, selector: #selector(update), userInfo: nil, repeats: true)
    }
    
    override func viewWillDisappear(_ animated: Bool) {
        timer.invalidate()
    }
    
    @objc func update() {
        if !isEditing {
            Transfers.refresh()
            tableView.reloadData()
        }
    }

    // MARK: - Table view data source

    override func numberOfSections(in tableView: UITableView) -> Int {
        return 3
    }

    override func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        switch section {
        case 0:
            return Transfers.inProgress.count
        case 1:
            return Transfers.completed.count
        case 2:
            return Transfers.failed.count
        default:
            return 0
        }
    }
    
    override func tableView(_ tableView: UITableView, titleForHeaderInSection section: Int) -> String? {
        if self.tableView(tableView, numberOfRowsInSection: section) > 0 {
            switch section {
            case 0:
                return "In progress"
            case 1:
                return "Completed"
            case 2:
                return "Failed"
            default:
                break
            }
        }
        return nil
        
    }

    override func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        var job: Job?
        switch indexPath.section {
        case 0:
            job = Transfers.inProgress[indexPath.row]
        case 1:
            job = Transfers.completed[indexPath.row]
        case 2:
            job = Transfers.failed[indexPath.row]
        default: break
        }
        let cell = tableView.dequeueReusableCell(withIdentifier: "reuseIdentifier", for: indexPath)
        if let cell = cell as? TransferTableViewCell {
            if job!.status?.success ?? false {
                cell.progressBar.isHidden = true
            } else if job!.status?.error != "" {
                cell.statusLabel.text = job?.status?.error
            } else {
                let bcf = ByteCountFormatter()
                cell.statusLabel.text = "\(bcf.string(fromByteCount: job?.stats?.bytes ?? 0)) of \(bcf.string(fromByteCount: job?.size ?? 0))"
            }
            cell.progressBar.setProgress(job?.percentDone ?? 0, animated: false)
            cell.nameLabel.text = job?.name
            cell.tag = job?.jobid ?? 0
        }
        return cell
    }
    
    // MARK: - Table view delegate
    
    override func tableView(_ tableView: UITableView, commit editingStyle: UITableViewCell.EditingStyle, forRowAt indexPath: IndexPath) {
        if editingStyle == .delete {
            //if transfer is in progress cancel it
            if (indexPath.section == 0) {
                Transfers.inProgress[indexPath.row].stop()
            }
            //Remove transfer from Transfers
            Transfers.remove(id: tableView.cellForRow(at: indexPath)!.tag)
            //Remove the row from tableview
            tableView.deleteRows(at: [indexPath], with: .fade)
        }
    }
}

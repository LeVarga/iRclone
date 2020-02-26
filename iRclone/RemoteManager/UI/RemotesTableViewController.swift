//
//  RemotesTableViewController.swift
//  iRclone
//
//  Created by Levente Varga on 2/18/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit
import Foundation

class RemotesTableViewController: UITableViewController {
    // MARK: - Properties
    
    var remotes = Remotes()
    var selectedRemote: String?

    //MARK: -
    
    override func viewDidLoad() {
        Rclone.setup()
        Rclone.start()
        super.viewDidLoad()
    }
    
    override func viewDidAppear(_ animated: Bool) {
        selectedRemote = nil
    }
    
    override func viewWillAppear(_ animated: Bool) {
        reload()
    }
    
    func reload() {
        self.remotes.load { (error) in
            self.tableView.reloadData()
        }
    }
    
    // MARK: - Table view data source
    
    //number of rows
    override func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return remotes.count
    }
    
    //cell setup
    override func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        let cell = UITableViewCell(style: .default, reuseIdentifier: "Cell")
        cell.textLabel?.text = remotes[indexPath.row].name! + " (\(remotes[indexPath.row].type)) "
        return cell
    }
    
    // MARK: - Table view delegate
    
    //row selected
    override func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        tableView.deselectRow(at: indexPath, animated: true)
        print("Remote selected: " + remotes[indexPath.row].name!)
        selectedRemote = remotes[indexPath.row].name! + ":"
        performSegue(withIdentifier: "selectRemote", sender: self)
    }
    
    //editing
    override func tableView(_ tableView: UITableView, commit editingStyle: UITableViewCell.EditingStyle, forRowAt indexPath: IndexPath) {
        if editingStyle == .delete {
            remotes.delete(remote: remotes[indexPath.row], completion: { (error) in
                self.reload()
            })
        }
    }
    
    // MARK: - Segue
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        if let theDestination = segue.destination as? RemoteFilesTableViewController {
            theDestination.remote = selectedRemote
        }
    }
}

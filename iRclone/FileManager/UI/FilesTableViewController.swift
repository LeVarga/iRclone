//
//  FilesTableViewController.swift
//  iRclone
//
//  Created by Levente Varga on 2/7/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit

class FilesTableViewController: UITableViewController {
    // MARK: - Properties
    var remote: String? //remote in which the working directory is located (nil if local)
    var wd: URL? //working directory
    var contents: [File] = [] { //contents of working directory
        didSet {
            tableView.reloadData()
        }
    }
    var pasteBtn: UIBarButtonItem?
    var copyBtn: UIBarButtonItem?
    var cutBtn: UIBarButtonItem?
    var mkdirBtn: UIBarButtonItem?
    
    // MARK: -
    override func viewDidLoad() {
        if wd == nil && remote == nil {
            wd = documentsUrl
        }
        //set up refresh control target and start loading contents
        self.refreshControl?.addTarget(self, action: #selector(reload), for: .valueChanged)
        self.refreshControl?.beginRefreshing()
        self.refreshControl?.sendActions(for: .valueChanged)

        self.navigationItem.rightBarButtonItem = self.editButtonItem
        
        //set up edit toolbar items
        copyBtn = UIBarButtonItem(image: UIImage(systemName: "doc.on.doc"), style: .plain, target: self, action: #selector(copyButton))
        copyBtn!.setTitleTextAttributes([NSAttributedString.Key.foregroundColor : UIColor.lightGray], for: .disabled)
        pasteBtn = UIBarButtonItem(image: UIImage(systemName: "doc.on.clipboard"), style: .plain, target: self, action: #selector(pasteButton))
        pasteBtn!.setTitleTextAttributes([NSAttributedString.Key.foregroundColor : UIColor.lightGray], for: .disabled)
        mkdirBtn = UIBarButtonItem(image: UIImage(systemName: "folder.badge.plus"), style: .plain, target: self, action: #selector(mkdirButton))
        mkdirBtn!.setTitleTextAttributes([NSAttributedString.Key.foregroundColor : UIColor.lightGray], for: .disabled)
        cutBtn = UIBarButtonItem(image: UIImage(systemName: "scissors"), style: .plain, target: self, action: #selector(cutButton))
        cutBtn!.setTitleTextAttributes([NSAttributedString.Key.foregroundColor : UIColor.lightGray], for: .disabled)
        let flexibleSpace = UIBarButtonItem(barButtonSystemItem: .flexibleSpace, target: self, action: nil)
        self.toolbarItems = [copyBtn!, cutBtn!, flexibleSpace, pasteBtn!, flexibleSpace, mkdirBtn!]
        self.navigationItem.title = wd?.lastPathComponent ?? remote
    }
    
    override func viewWillAppear(_ animated: Bool) {
        self.navigationController?.setToolbarHidden(true, animated: true)
        self.setEditing(false, animated: false)
    }
    
    override func setEditing(_ editing: Bool, animated: Bool) {
        super.setEditing(editing, animated: animated)
        self.navigationController?.setToolbarHidden(!editing, animated: true)
        refreshToolbarButtons()
    }
    
    func refreshToolbarButtons() {
        let condition = tableView.indexPathsForSelectedRows?.count ?? 0 > 0
        pasteBtn?.isEnabled = clipboard != nil && !condition
        mkdirBtn?.isEnabled = !condition
        copyBtn?.isEnabled = condition
        cutBtn?.isEnabled = condition
    }
    
    // MARK: - Table view data source
    override func numberOfSections(in tableView: UITableView) -> Int {
        return 1
    }
    
    override func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return contents.count
    }
    
    override func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        let cell = tableView.dequeueReusableCell(withIdentifier: "reuseIdentifier", for: indexPath)
        cell.textLabel?.text = contents[indexPath.row].name
        
        //set cell image based on file type
        var imageName: String
        switch contents[indexPath.row].type {
        case .video: imageName = "film"
        case .image: imageName = "photo"
        case .directory: imageName = "folder"
        case .document: imageName = "doc"
        case .audio: imageName = "music.note"
        default: imageName = "questionmark"
        }
        cell.imageView?.image = UIImage(systemName: imageName)
        
        return cell
    }
    
    // MARK: - Table view delegate
    override func tableView(_ tableView: UITableView, didDeselectRowAt indexPath: IndexPath) {
        refreshToolbarButtons()
    }
    
    override func tableView(_ tableView: UITableView, commit editingStyle: UITableViewCell.EditingStyle, forRowAt indexPath: IndexPath) {
        let item = contents[indexPath.row]
        if editingStyle == .delete {
            let alert = UIAlertController(title: "Delete \(item.isDir ? "directory" : "file")", message: "Are you sure you want to delete \(item.path)?", preferredStyle: UIAlertController.Style.alert)
            alert.addAction(UIAlertAction(title: "Delete", style: .destructive, handler: { _ in
                item.delete(remote: self.remote) { (error) in
                    if let err = error {
                        self.presentError(error: err)
                    } else {
                        self.reload()
                    }
                }
                }))
            alert.addAction(UIAlertAction(title: "Cancel", style: .cancel))
            self.present(alert, animated: true, completion: nil)
        }
    }
    
    // MARK: - Actions
    @objc func reload() {
        self.refreshControl?.beginRefreshing()
        if remote != nil { //directory is remote
            RemoteFiles.load(remote: self.remote!, path: self.wd, completion: {(error, files) in
                self.refreshControl!.endRefreshing()
                if let err = error {
                    self.presentError(error: err) {_ in
                        self.dismiss(animated: true)
                    }
                } else {
                    self.contents = files.sorted(by: {$0.name.lowercased() < $1.name.lowercased()})
                }
            })
        } else { //directory is local
            if let files = try? FileManager.default.contentsOfDirectory(at: wd ?? documentsUrl, includingPropertiesForKeys: [.fileSizeKey]) {
                var tmp: [LocalFile] = []
                files.forEach { (file) in
                    tmp.append(LocalFile(file: file))
                }
                self.contents = tmp.sorted(by: {
                    $0.name.lowercased() < $1.name.lowercased()
                })
            }
            self.refreshControl?.endRefreshing()
        }
    }
    
    @objc func mkdirButton() {
        presentInputDialog(title: "New folder", subtitle: nil, actionTitle: "Create", cancelTitle: "Cancel", inputPlaceholder: "Untitled folder", inputKeyboardType: .default, cancelHandler: { _ in }) { (name) in
            let newDirURL = URL(fileURLWithPath: (self.wd?.path ?? "") + "/\(name ?? "Untitled folder")/", isDirectory: true)
            if self.remote != nil {
                RemoteFiles.mkdir(remote: self.remote!, path: newDirURL) { (error) in
                    if error != nil {
                        self.presentError(error: error!)
                        return
                    }
                }
            } else {
                do {
                    try FileManager.default.createDirectory(at: newDirURL, withIntermediateDirectories: false, attributes: nil)
                } catch {
                    self.presentError(error: error)
                }
            }
            self.viewWillAppear(true)
            self.reload()
        }
    }
    
    @objc func pasteButton() {
        // TODO: Display errors
        let _ = clipboard?.paste(dstPath: wd?.path, dstRemoteFs: remote)
        viewWillAppear(true)
     }
    
    @objc func cutButton() {
        if let selectedIndexPaths = self.tableView.indexPathsForSelectedRows {
            clipboard = Clipboard(files: [], move: true, srcRemoteFs: remote)
            for indexPath in selectedIndexPaths {
                clipboard!.files.append(contents[indexPath.row])
            }
        }
        viewWillAppear(true)
    }
    
    @objc func copyButton() {
        if let selectedIndexPaths = self.tableView.indexPathsForSelectedRows {
            clipboard = Clipboard(files: [], move: false, srcRemoteFs: remote)
            for indexPath in selectedIndexPaths {
                clipboard!.files.append(contents[indexPath.row])
            }
        }
        viewWillAppear(true)
    }
}

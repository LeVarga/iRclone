//
//  RemoteFilesTableViewController.swift
//  iRclone
//
//  Created by Levente Varga on 2/7/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit
import QuickLook

class RemoteFilesTableViewController: FilesTableViewController, QLPreviewControllerDataSource, QLPreviewControllerDelegate {
    // MARK: - Properties
    var selectedUrl: URL?
    lazy var previewItem = NSURL()
    var previewDownloadSession: URLSessionDataTask?
    
    // MARK: - Preview Controller data source
    func numberOfPreviewItems(in controller: QLPreviewController) -> Int {
        return 1
    }

    func previewController(_ controller: QLPreviewController, previewItemAt index: Int) -> QLPreviewItem {
        return previewItem
    }
    
    // MARK: - Preview Controller delegate
    func previewControllerWillDismiss(_ controller: QLPreviewController) {
        previewDownloadSession?.cancel()
    }
    
    // MARK: -
    override func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        if !tableView.isEditing {
            tableView.deselectRow(at: indexPath, animated: true)
            let selected = contents[indexPath.row]
            selectedUrl = URL(string: Rclone.rcUrl.absoluteString + "/[\(self.remote!)]/\(selected.path)".addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)!)
            switch selected.type {
            case .directory:
                let vc = storyboard?.instantiateViewController(withIdentifier: "remoteFiles")
                if let vc = vc as? RemoteFilesTableViewController {
                    vc.wd = URL(string: selected.path.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)!)
                    vc.remote = self.remote
                    self.navigationController?.pushViewController(vc, animated:true)
                }
            case .video:
                performSegue(withIdentifier: "VideoPlayer", sender: self)
            default:
                let quickLookController = QLPreviewController()
                quickLookController.dataSource = self
                quickLookController.currentPreviewItemIndex = 0
                let tmpUrl = FileManager().temporaryDirectory.appendingPathComponent(selected.name)
                let indicator = UIActivityIndicatorView(style: .large)
                if ((try? tmpUrl.checkResourceIsReachable()) ?? false), (Int64((try! tmpUrl.resourceValues(forKeys:[.fileSizeKey])).fileSize!) == selected.size) {
                    previewItem = tmpUrl as NSURL
                } else {
                    indicator.center = CGPoint(x: self.view.frame.width / 2.0, y: self.view.frame.width / 2.0)
                    quickLookController.view.addSubview(indicator)
                    quickLookController.view.bringSubviewToFront(indicator)
                    indicator.startAnimating()
                    previewDownloadSession = URLSession.shared.dataTask(with: selectedUrl!) { data, response, error in
                        guard let data = data, error == nil else {
                            DispatchQueue.main.async {
                                self.presentError(error: error!) { _ in
                                    quickLookController.dismiss(animated: true)
                                }
                            }
                            return
                        }
                        do {
                            try data.write(to: tmpUrl, options: .atomic)   // atomic option overwrites it if needed
                            DispatchQueue.main.async {
                                indicator.stopAnimating()
                                self.previewItem = tmpUrl as! NSURL
                                quickLookController.refreshCurrentPreviewItem()
                            }
                        } catch {
                            self.presentError(error: error) { _ in
                                quickLookController.dismiss(animated: true)
                            }
                            return
                        }
                        
                    }
                    previewDownloadSession?.resume()
                }
                self.present(quickLookController, animated: true)
            }
        } else {
            refreshToolbarButtons()
        }
    }
    
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        if let destVideoVC = segue.destination as? VideoViewController {
            destVideoVC.url = self.selectedUrl
        }
    }
}

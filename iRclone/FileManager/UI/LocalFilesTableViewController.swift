//
//  LocalFilesTableViewController.swift
//  iRclone
//
//  Created by Levente Varga on 2/7/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit
import QuickLook

class LocalFilesTableViewController: FilesTableViewController, QLPreviewControllerDataSource {
    var selectedURL: URL?
    
    // MARK: - Preview Controller data source
    func numberOfPreviewItems(in controller: QLPreviewController) -> Int {
        1
    }
    
    func previewController(_ controller: QLPreviewController, previewItemAt index: Int) -> QLPreviewItem {
        self.selectedURL! as QLPreviewItem
    }
    
    // MARK: - Table view delegate
    override func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        if !tableView.isEditing {
            tableView.deselectRow(at: indexPath, animated: true)
            let selected = contents[indexPath.row]
            selectedURL = URL(string: "file://" + selected.path.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)!)
            switch selected.type {
            case .directory:
                let vc = storyboard?.instantiateViewController(withIdentifier: "localFiles")
                if let vc = vc as? LocalFilesTableViewController {
                    vc.wd = URL(string: selected.path)!
                    self.navigationController?.pushViewController(vc, animated:true)
                }
            case .video: performSegue(withIdentifier: "VideoPlayer", sender: self)
            case .image: preview()
            case .audio: preview()
            case .document: preview()
            default: break
            }
        } else {
            refreshToolbarButtons()
        }
    }
    
    // MARK: -
    ///Open the last selected file in QLPreviewController
    func preview() {
        let quickLookController = QLPreviewController()
        quickLookController.dataSource = self
        if QLPreviewController.canPreview(selectedURL! as QLPreviewItem) {
            quickLookController.currentPreviewItemIndex = 0
            present(quickLookController, animated: true, completion: nil)
        }
    }
    
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        if let dest = segue.destination as? VideoViewController {
            dest.path = selectedURL?.path
        }
    }
}

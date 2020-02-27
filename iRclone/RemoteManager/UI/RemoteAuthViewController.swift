//
//  RemoteAuthViewController.swift
//  iRclone
//
//  Created by Levente Varga on 1/4/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit
import WebKit
import Rclone

class RemoteAuthViewController: UIViewController {
    var name: String?
    
    @IBOutlet var webView: WKWebView!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        let request = URLRequest(url: URL(string: "http://localhost:53682/auth?state=" + RcloneGetAuthState())!)
        
        webView.load(request)
        
        // Do any additional setup after loading the view.
    }
    @IBAction func cancelButtonAction(_ sender: Any) {
        if let url = URL(string: ("http://localhost:53682/?state=" + RcloneGetAuthState() + "&code=cancel")) {
                   var request = URLRequest(url: url)
                   request.httpMethod = "POST"
                   let task = URLSession.shared.dataTask(with: request)
            self.presentingViewController?.dismiss(animated: false, completion: {
                task.resume()
            })
               }
    }
    
    /*
    // MARK: - Navigation

    // In a storyboard-based application, you will often want to do a little preparation before navigation
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        // Get the new view controller using segue.destination.
        // Pass the selected object to the new view controller.
    }
    */

}

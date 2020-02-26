//
//  NewRemoteViewController.swift
//  iRclone
//
//  Created by Levente Varga on 6/8/18.
//  Copyright © 2018 Levente V. All rights reserved.
//

import Foundation
import UIKit

class NewRemoteViewController: UIViewController, UIPickerViewDelegate, UIPickerViewDataSource, UITextFieldDelegate {
    // MARK: - Properties
    
    @IBOutlet weak var nameField: UITextField!
    @IBOutlet var providerPicker: UIPickerView! {
        didSet {
            providerPicker.delegate = self
            providerPicker.dataSource = self
        }
    }
    @IBOutlet var remoteNameTextField: UITextField!
    var providerSelected: Int {
        get {
            providerPicker.selectedRow(inComponent: 0)
        }
    }
    var providers: [Provider]?
    
    // MARK: -
    override func viewDidLoad() {
        super.viewDidLoad()
        Rclone.request(queryString: "config/providers", jsonData: nil, timeout: 5, decodeAs: Providers.self) { (decoded, error) in
            if let providers = decoded?.providers {
                self.providers = providers
            }
            self.providerPicker.reloadAllComponents()
        }
    }
    
    // MARK: - Picker view data source
    func numberOfComponents(in pickerView: UIPickerView) -> Int {
        return 1
    }

    func pickerView(_ pickerView: UIPickerView, numberOfRowsInComponent component: Int) -> Int {
        return providers?.count ?? 0
    }
    
    func pickerView(_ pickerView: UIPickerView, titleForRow row: Int, forComponent component: Int) -> String? {
        return providers?[row].description
    }
    
    // MARK: - Segue
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        if let optionsVC = segue.destination as? RemoteOptionsViewController {
            optionsVC.allOptions = providers![providerSelected].options ?? []
            optionsVC.providerName = providers![providerSelected].name
            optionsVC.chosenName = remoteNameTextField.text!
        }
    }
    
    override func shouldPerformSegue(withIdentifier identifier: String, sender: Any?) -> Bool {
        if (identifier == "Options" && remoteNameTextField.text?.trimmingCharacters(in: CharacterSet.whitespacesAndNewlines) != "") {
            return true
        }
        print("Name field empty or invalid segue id")
        return false
    }
}

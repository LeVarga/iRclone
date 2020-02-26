//
//  RemoteOptionsViewController.swift
//  iRclone
//
//  Created by Levente Varga on 1/3/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit
import WebKit

class RemoteOptionsViewController: UIViewController, UITextFieldDelegate, UITableViewDelegate, UITableViewDataSource {
    // MARK: - Properties
    
    @IBOutlet var prototypeHelpButton: UIButton!
    @IBOutlet var prototypeTextField: UITextField!
    
    @IBOutlet var prototypeLabel: UILabel!
    @IBOutlet var contentView: UIView!
    
    @IBOutlet var tableView: UITableView! {
        didSet {
            self.tableView.delegate = self
            self.tableView.dataSource = self
        }
    }
    var chosenName = ""
    var providerName = ""
    var allOptions = [ProviderOptions]()
    var optionsBySection = [Section: [ProviderOptions]]()
    var cells = [UITableViewCell]()
    var params: [String: Any] = [:]
    
    //MARK: -
   
    override func viewDidLoad() {
        if allOptions.count == 0 {
            nextButtonAction(UIButton())
        }
        optionsBySection[.Required] = allOptions.filter { (option) -> Bool in
            return option.optionRequired
        }
        optionsBySection[.Optional] = allOptions.filter { (option) -> Bool in
            return (!option.advanced && !option.optionRequired)
        }
        optionsBySection[.Advanced] = allOptions.filter { (option) -> Bool in
            return (option.advanced && !option.optionRequired)
        }
        super.viewDidLoad()
    }
    
    //MARK: - Table view data source
    enum Section: Int {
        case Required = 0
        case Optional = 1
        case Advanced = 2
    }
    
    func numberOfSections(in tableView: UITableView) -> Int {
        return optionsBySection.count
    }
    
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return optionsBySection[Section(rawValue: section)!]?.count ?? 0
    }
    
    func tableView(_ tableView: UITableView, titleForHeaderInSection section: Int) -> String? {
        if section == Section.Required.rawValue && (optionsBySection[.Required]?.count ?? 0) > 0 {
            return "Required"
        } else if section == Section.Optional.rawValue && (optionsBySection[.Optional]?.count ?? 0) > 0 {
            return "Optional"
        } else if section == Section.Advanced.rawValue && (optionsBySection[.Advanced]?.count ?? 0) > 0 {
            return "Advanced"
        }
        return nil
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        var option: ProviderOptions?
        option = optionsBySection[Section(rawValue: indexPath.section)!]?[indexPath.row]
        let cell: UITableViewCell
        if option?.type == "bool" {
            cell = tableView.dequeueReusableCell(withIdentifier: "BoolOptionCell")!
            if let cell = cell as? BoolOptionTableViewCell {
                cell.callback = { sw -> Void in
                    self.params[option!.name] = cell.optionSwitch.isOn
                }
                cell.optionNameLabel.text = option?.name
                cell.optionSwitch.isOn = self.params[option!.name] as? Bool ?? (option?.defaultStr == "true")
            }
        } else {
            cell = tableView.dequeueReusableCell(withIdentifier: "StringOptionCell")!
            if let cell = cell as? StringOptionTableViewCell {
                cell.optionNameLabel.text = option?.name
                cell.optionValueTextField.placeholder = option?.defaultStr
                cell.callback = { tf -> Void in
                    if let value = tf.text?.trimmingCharacters(in: .whitespacesAndNewlines), value != "" {
                        self.params[option!.name] = value
                    } else {
                        self.params.removeValue(forKey: option!.name)
                    }
                }
                cell.optionValueTextField.text = self.params[option!.name] as? String
                cell.optionValueTextField.isSecureTextEntry = option?.isPassword ?? false
                cell.optionValueTextField.keyboardType = option?.type == "int" ? .numberPad : .default
            }
        }
        cells.append(cell)
        return cell
    }
    
    //MARK: - Table view delegate
    
    func tableView(_ tableView: UITableView, accessoryButtonTappedForRowWith indexPath: IndexPath) {
        let option = optionsBySection[Section(rawValue: indexPath.section)!]![indexPath.row]
        var message = "\(option.help)"
         if option.examples?.count ?? 0 > 0 {
             message += "\nExamples: "
             for example in option.examples! {
                 message += "\n" + example.value!
                 message += " (" + (example.help ?? "No description") + ")"
             }
         }
         let alert = UIAlertController(title: option.name, message: message, preferredStyle: UIAlertController.Style.alert)
         alert.addAction(UIAlertAction(title: "Dismiss", style: .default, handler: nil))
         self.present(alert, animated: true, completion: nil)
         print(option.help)
    }
    
    //MARK: - Segue
    
    override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
        if segue.identifier == "Authorization" {
            let vc = segue.destination as! RemoteAuthViewController
            vc.name = chosenName
        }
    }
    
    //MARK: - Action funcs
    
    @IBAction func nextButtonAction(_ sender: UIButton) {
        let json: [String: Any] = ["name": chosenName,
                                   "type": providerName,
                                   "parameters": params]
        print(json)
        let jsonData = try? JSONSerialization.data(withJSONObject: json)
        Rclone.request(queryString: "config/create", jsonData: jsonData, timeout: .infinity, decodeAs: Empty.self, completion: { _, error in
            if let error = error {
                self.presentError(error: error) { _ in
                    self.dismiss(animated: true)
                }
            }
            if (self.presentedViewController != nil) {
                self.dismiss(animated: true) {
                    self.navigationController?.popToRootViewController(animated: true)
                }
            }
        })
        sleep(1)
        if Rclone.authState != "" {
            performSegue(withIdentifier: "Authorization", sender: self)
        }
        
    }
}

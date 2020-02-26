//
//  StringOptionTableViewCell.swift
//  iRclone
//
//  Created by Levente Varga on 2/19/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit

class StringOptionTableViewCell: UITableViewCell {
    @IBOutlet var optionNameLabel: UILabel!
    var callback: ((_ textfield: UITextField) -> Void)?
    @IBOutlet var optionValueTextField: UITextField!
    override func awakeFromNib() {
        super.awakeFromNib()
        // Initialization code
    }

    override func setSelected(_ selected: Bool, animated: Bool) {
        super.setSelected(selected, animated: animated)
        // Configure the view for the selected state
    }

    @IBAction func editingChanged(_ sender: UITextField) {
        callback?(sender)
    }
}

//
//  BoolOptionTableViewCell.swift
//  iRclone
//
//  Created by Levente Varga on 2/19/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit

class BoolOptionTableViewCell: UITableViewCell {
    @IBOutlet var optionSwitch: UISwitch!
    @IBOutlet var optionNameLabel: UILabel!
    override func awakeFromNib() {
        super.awakeFromNib()
        // Initialization code
    }
    
    var callback: ((_ switch: UISwitch) -> Void)?

    @IBAction func switchValueChanged(_ sender: UISwitch) {
        callback?(sender)
    }
    override func setSelected(_ selected: Bool, animated: Bool) {
        super.setSelected(selected, animated: animated)

        // Configure the view for the selected state
    }

}

//
//  TransferTableViewCell.swift
//  iRclone
//
//  Created by Levente Varga on 2/11/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

import UIKit

class TransferTableViewCell: UITableViewCell {
    @IBOutlet var nameLabel: UILabel!
    @IBOutlet var statusLabel: UILabel!
    @IBOutlet var progressBar: UIProgressView!
    
    override func awakeFromNib() {
        super.awakeFromNib()
        // Initialization code
    }

    override func setSelected(_ selected: Bool, animated: Bool) {
        super.setSelected(selected, animated: animated)

        // Configure the view for the selected state
    }

}

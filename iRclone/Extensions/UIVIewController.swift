//
//  UIVIewController.swift
//  iRclone
//
//  Created by Levente Varga on 2/17/20.
//  Copyright © 2020 Levente V. All rights reserved.
//

extension UIViewController {
    func presentError(error: String, cancelHandler: ((UIAlertAction) -> Swift.Void)? = nil) {
        let alert = UIAlertController(title: "Error", message: error, preferredStyle: UIAlertController.Style.alert)
        alert.addAction(UIAlertAction(title: "Cancel", style: .cancel, handler: cancelHandler))
        self.present(alert, animated: true, completion: nil)
    }
    
    func presentError(error: Error, cancelHandler: ((UIAlertAction) -> Swift.Void)? = nil) {
        presentError(error: error.localizedDescription, cancelHandler: cancelHandler)
    }
    
    func presentInputDialog(title: String? = nil,
                         subtitle: String? = nil,
                         actionTitle: String? = "Add",
                         cancelTitle: String? = "Cancel",
                         inputPlaceholder: String? = nil,
                         inputKeyboardType: UIKeyboardType = UIKeyboardType.default,
                         cancelHandler: ((UIAlertAction) -> Swift.Void)? = nil,
                         actionHandler: ((_ text: String?) -> Void)? = nil) {

        let alert = UIAlertController(title: title, message: subtitle, preferredStyle: .alert)
        alert.addTextField { (textField:UITextField) in
            textField.placeholder = inputPlaceholder
            textField.keyboardType = inputKeyboardType
        }
        alert.addAction(UIAlertAction(title: actionTitle, style: .destructive, handler: { (action:UIAlertAction) in
            guard let textField =  alert.textFields?.first else {
                actionHandler?(nil)
                return
            }
            actionHandler?(textField.text)
        }))
        alert.addAction(UIAlertAction(title: cancelTitle, style: .cancel, handler: cancelHandler))

        self.present(alert, animated: true, completion: nil)
    }
}

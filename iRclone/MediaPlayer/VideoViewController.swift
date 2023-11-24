//
//  VideoViewController.swift
//  iRclone
//
//  Created by Levente Varga on 11/20/19.
//  Copyright © 2019 Levente V. All rights reserved.
//

import UIKit
import MobileVLCKit
import QuartzCore

class VideoViewController: UIViewController, VLCMediaPlayerDelegate, UITableViewDelegate, UITableViewDataSource {
    // MARK: - Properties
    
    lazy var videoView: UIView! = {
        let vw = UIView()
        vw.backgroundColor = UIColor.black
        let gesture = UITapGestureRecognizer(target: self, action: #selector(self.movieViewTapped(_:)))
        vw.addGestureRecognizer(gesture)
        return vw
    }()
    let spinner = SpinnerViewController()
    
    var controlsVisible = false
    var seekBarBeingTouched = false
    var url: URL?
    var path: String?
    var mediaPlayer: VLCMediaPlayer = VLCMediaPlayer()
    
    @IBOutlet var controlsView: UIVisualEffectView! {
        didSet {
            controlsView.clipsToBounds = true
            controlsView.layer.cornerRadius = 25
        }
    }
    @IBOutlet var remainingTimeLabel: UILabel!
    @IBOutlet var currentTimeLabel: UILabel!
    @IBOutlet weak var seekBar: UISlider!
    @IBOutlet var playPauseButton: UIButton!
    @IBOutlet weak var exitButton: UIButton!
    @IBOutlet var trackSelectionSubView: UIView! // TODO: Make track menu work in landscape and look less bad
    @IBOutlet var tracksTableView: UITableView! {
        didSet {
            tracksTableView.dataSource = self
            tracksTableView.delegate = self
        }
    }
    
    // MARK: - Player setup
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        //add videoView subview
        self.view.addSubview(self.videoView)
    }
    
    override func viewDidAppear(_ animated: Bool) {
        assert(url != nil || path != nil, "Missing media url/path")

        var media: VLCMedia
        self.videoView.frame = self.view.bounds
        if let url = url {
            media = VLCMedia(url: url)
            toggleSpinner(on: true)
            print(url)
        } else {
            media = VLCMedia(path: path!)
        }
        media.addOptions([
            "network-caching": 1000
        ])
        
        mediaPlayer.media = media
        mediaPlayer.delegate = self
        mediaPlayer.drawable = self.videoView
        mediaPlayer.play()
    }
    
    override func viewWillTransition(to size: CGSize, with coordinator: UIViewControllerTransitionCoordinator) {
        //self.videoView.frame = UIScreen.screens[0].bounds
        self.videoView.frame = CGRect(origin: CGPointZero, size: size)
    }
    
    // MARK: - Table view data source
    
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        switch section {
        case 0:
            return Int(mediaPlayer.numberOfAudioTracks)
        case 1:
            return Int(mediaPlayer.numberOfSubtitlesTracks)
        case 2:
            return Int(mediaPlayer.numberOfVideoTracks)
        default:
            return 0
        }
    }
    
    func numberOfSections(in tableView: UITableView) -> Int {
        return 3
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        let cell = tableView.dequeueReusableCell(withIdentifier: "reuseIdentifier", for: indexPath)
        cell.accessoryType = .none
        switch indexPath.section {
        case 0:
            if mediaPlayer.audioTrackIndexes[indexPath.row] as? Int32 == mediaPlayer.currentAudioTrackIndex {
                cell.accessoryType = .checkmark
            }
            if let audioTrackName = mediaPlayer.audioTrackNames[indexPath.row] as? String {
                cell.textLabel?.text = audioTrackName
                break
            }
            cell.textLabel?.text = "Track #\(indexPath.row + 1)"
        case 1:
            if mediaPlayer.videoSubTitlesIndexes[indexPath.row] as? Int32 == mediaPlayer.currentVideoSubTitleIndex {
                cell.accessoryType = .checkmark
            }
            if let subTitleTrackName = mediaPlayer.videoSubTitlesNames[indexPath.row] as? String {
                cell.textLabel?.text = subTitleTrackName
                break
            }
            cell.textLabel?.text = "Track #\(indexPath.row + 1)"
        case 2:
            if mediaPlayer.videoTrackIndexes[indexPath.row] as? Int32 == mediaPlayer.currentVideoTrackIndex {
                cell.accessoryType = .checkmark
            }
            if let videoTrackName = mediaPlayer.videoTrackNames[indexPath.row] as? String {
                cell.textLabel?.text = videoTrackName
                break
            }
            cell.textLabel?.text = "Track #\(indexPath.row + 1)"
        default: break
        }
        return cell
    }
    
    func tableView(_ tableView: UITableView, titleForHeaderInSection section: Int) -> String? {
        switch section {
        case 0:
            return "Audio"
        case 1:
            return "Subtitle"
        case 2:
            return "Video"
        default:
            return nil
        }
    }
    
    // MARK: - Table view delegate
    
    func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        switch indexPath.section {
        case 0:
            mediaPlayer.currentAudioTrackIndex = mediaPlayer.audioTrackIndexes[indexPath.row] as! Int32
        case 1:
            mediaPlayer.currentVideoSubTitleIndex = mediaPlayer.videoSubTitlesIndexes[indexPath.row] as! Int32
        case 2:
            mediaPlayer.currentVideoTrackIndex = mediaPlayer.videoTrackIndexes[indexPath.row] as! Int32
        default:
            break
        }
        tableView.reloadSections([indexPath.section], with: .none)
    }
    
    func numberOfComponents(in pickerView: UIPickerView) -> Int {
        return 1
    }
    
    //MARK: - VLCMediaPlayerDelegate
    
    func mediaPlayerTimeChanged(_ aNotification: Notification) {
        toggleSpinner(on: false)
        if controlsVisible && !seekBarBeingTouched {
            seekBar.setValue(mediaPlayer.time.value?.floatValue ?? 0, animated: false)
            currentTimeLabel.text = mediaPlayer.time.stringValue
            remainingTimeLabel.text = mediaPlayer.remainingTime?.stringValue
        }
    }
      
    func mediaPlayerStateChanged(_ aNotification: Notification) {
        //.state is broken, using mediaPlayerTimeChanged for detecting when playback starts instead.
        switch mediaPlayer.state {
        case .error: exitPlayer(0)
        default: break
        }
    }
    
    // MARK: - Action funcs
    
    @IBAction func fastForward15s(_ sender: Any) {
        mediaPlayer.jumpForward(15)
    }
    
    @objc func movieViewTapped(_ sender: UITapGestureRecognizer) {
        toggleControlsVisible()
        if mediaPlayer.isSeekable {
            seekBar.maximumValue = mediaPlayer.media?.length.value?.floatValue ?? 0
            seekBar.setValue(mediaPlayer.time.value?.floatValue ?? 0, animated: false)
        }
    }
    
    @IBAction func seekBarValueChanged(_ sender: Any) {
        mediaPlayer.time = VLCTime(number: NSNumber(floatLiteral: Double(seekBar.value)))
        if !mediaPlayer.isPlaying {
            mediaPlayer.gotoNextFrame()
        }
    }
    
    @IBAction func playPause(_ sender: Any) {
        if mediaPlayer.isPlaying && mediaPlayer.canPause {
            mediaPlayer.pause()
        }
        else if mediaPlayer.isPlaying == false {
            mediaPlayer.play()
        }
    }
    
    @IBAction func exitPlayer(_ sender: Any) {
        mediaPlayer.stop()
        mediaPlayer.media = nil
        dismiss(animated: false, completion: nil)
    }
       
    @IBAction func toggleTracksMenuVisible(_ sender: Any) {
        if (trackSelectionSubView.isHidden) {
            tracksTableView.reloadData()
        }
        trackSelectionSubView.isHidden.toggle()
    }
    
    @IBAction func seekBarTouchDown(_ sender: Any) {
        seekBarBeingTouched = true
    }
    @IBAction func seekBarTouchUp(_ sender: Any) {
        seekBarBeingTouched = false
    }
    
    @IBAction func rewind15s(_ sender: Any) {
        mediaPlayer.jumpBackward(15)
    }
    
    //MARK: -
    
    func toggleSpinner(on: Bool) {
        if on {
            addChild(spinner)
            spinner.view.frame = view.frame
            view.addSubview(spinner.view)
            spinner.didMove(toParent: self)
        } else {
            spinner.willMove(toParent: nil)
            spinner.view.removeFromSuperview()
            spinner.removeFromParent()
        }
    }
    
    func toggleControlsVisible() {
        if controlsVisible {
            self.view.sendSubviewToBack(exitButton)
            self.view.sendSubviewToBack(controlsView)
            self.view.sendSubviewToBack(trackSelectionSubView)
            trackSelectionSubView.isHidden = true
        } else {
            self.view.bringSubviewToFront(controlsView)
            self.view.bringSubviewToFront(exitButton)
            self.view.bringSubviewToFront(trackSelectionSubView)
        }
        controlsVisible.toggle()
        setNeedsUpdateOfHomeIndicatorAutoHidden()
    }
    
    class SpinnerViewController: UIViewController {
        var spinner = UIActivityIndicatorView(style: .whiteLarge)

        override func loadView() {
            view = UIView()
            view.backgroundColor = UIColor(white: 0, alpha: 0.7)

            spinner.translatesAutoresizingMaskIntoConstraints = false
            spinner.startAnimating()
            view.addSubview(spinner)

            spinner.centerXAnchor.constraint(equalTo: view.centerXAnchor).isActive = true
            spinner.centerYAnchor.constraint(equalTo: view.centerYAnchor).isActive = true
        }
    }
    
    override var prefersHomeIndicatorAutoHidden: Bool {
        return !controlsVisible
    }
}


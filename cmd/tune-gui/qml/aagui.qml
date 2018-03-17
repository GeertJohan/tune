import QtQuick 2.8
import QtQuick.Controls 2.1
import QtQuick.Layouts 1.3
import QtQuick.Window 2.2
import QtQuick.Templates 2.1

ApplicationWindow {
	id: window

	visible: true
	title: "AudioAddict Player"
	minimumWidth: 400
	minimumHeight: 400

	ColumnLayout {
		id: columnLayout
		anchors.rightMargin: 0
		anchors.bottomMargin: 0
		anchors.leftMargin: 0
		anchors.topMargin: 0
		anchors.fill: parent

		ToolBar {
			id: toolBar
			width: 360
			Layout.minimumHeight: 50
			Layout.fillHeight: false
			Layout.fillWidth: true
		}

		RowLayout {
			id: mainView
			width: 100
			height: 100
			Layout.fillHeight: true
			Layout.fillWidth: true


			ListView {
				id: channelList
				x: 0
				y: 0
				width: 110
				height: 160
				Layout.fillHeight: true
				delegate: Item {
					id: item1
					x: 0
					width: 80
					height: 44
					Row {
						id: row1
						anchors.fill: parent
						anchors.topMargin: 2
						anchors.bottomMargin: 2

						Image {
							id: image
							anchors.top: parent.top
							anchors.bottom: parent.bottom
							fillMode: Image.PreserveAspectFit
							source: channelImage
						}

						Text {
							text: name
							font.bold: true
							anchors.verticalCenter: parent.verticalCenter
						}

						spacing: 10
					}
				}
				model: ListModel {
					ListElement {
						name: "Techno"
						channelImage: "https://cdn-images.audioaddict.com/6/b/2/4/1/b/6b241b9d1fc2680a070c81739413b028.jpg?size=72x72"
					}
					ListElement {
						name: "Minimal"
						channelImage: "https://cdn-images.audioaddict.com/9/2/3/b/4/6/923b46e62b92426918ed416246c26f6b.jpg?size=64x64"
					}
				}
				
				function addChannel(channelKey, channelName, channelImage, trackTitle) {
					channelList.model.append({
						name: channelName,
						channelImage: channelImage
					});
				}
				
				Connections {
					target: channelBridge
					onClearChannels: {
						channelList.model.clear();
					}
					onAddChannel: channelList.addChannel(channelKey, channelName, channelImage, trackTitle)
				}
			}

			ColumnLayout {
				id: channelPlayer
				width: 100
				height: 100
				Layout.fillHeight: true
				Layout.fillWidth: true

				RowLayout {
					id: channelDetails
					width: 100
					height: 100
					Layout.fillWidth: true
					Layout.fillHeight: true
				}

				Item {
					id: playerInteractions
					width: 200
					height: 200
					Layout.preferredHeight: 0
					Layout.minimumHeight: 50
					Layout.fillWidth: true

				}

				Item {
					id: playerInfo
					width: 200
					height: 200
					Layout.alignment: Qt.AlignLeft | Qt.AlignBottom
					Layout.fillWidth: true
					Layout.maximumHeight: 50
				}

			}

		}
	}
}
